#!/usr/bin/env perl
# cursor-agent に Claude Code の Scroll と同じ Alt+J/K/U/D を足す。
# Alt+J/K は 1 行、Alt+D/U は 1 ページ。CLI にキーバインド設定が無いためバンドルを書き換える。
# Herdr は Codex 以外へ Alt キーをそのまま渡すので、受け側だけを変える。
use strict;
use warnings;

# 会話ビューのスクロール。入力無効の return の直後、プロンプト中の矢印処理の前。
# プロンプトがアクティブでも j/k/u/d を文字入力にしない位置は下の入力欄側。
my $scroll_re =
qr/if\((\w+)\)return;if\((\w+)\)return\(0,\w+\.Po\)\((\w+),(\w+)\)\?void (\w+)\(-1\):\(0,\w+\.u\$\)\(\3,\4\)\?void \5\(1\):\4\.ctrl&&"f"===\3\?void \5\(Math\.max\(1,(\w+)-1\)\):\4\.ctrl&&"b"===\3\?void \5\(-Math\.max\(1,\6-1\)\):\4\.ctrl&&"d"===\3\?void \5\(Math\.max\(1,Math\.floor\(\6\/2\)\)\):\4\.ctrl&&"u"===\3\?void \5\(-Math\.max\(1,Math\.floor\(\6\/2\)\)\):void 0;/;

# 入力欄が同じ Alt キーを文字や単語削除として扱わないようにする。
# 先頭の disable 判定の直後。bracketed paste のフィルタがあるコールバックだけに当たる。
my $input_re =
qr/\.Ge\)\(\(\((\w+),(\w+)\)=>\{(?:var (?:\w+,)*\w+;)?if\((\w+)\)return;(?=const \w+=\w+\.current\(\1\);if\(0===\w+\.length&&\1\.length>0&&!\w+\(\2\)\)return;)/;

my $scroll_n = 0;
my $input_n  = 0;

for my $file (@ARGV) {
    open my $in, '<', $file or die "cannot read $file: $!";
    my $src = do { local $/; <$in> };
    close $in;

    my $n_scroll = 0;
    $src =~ s{$scroll_re}{
        my ( $disabled, $prompt, $input, $key, $scroll, $height ) =
          ( $1, $2, $3, $4, $5, $6 );
        my $rest = $&;
        $rest =~ s/^if\(\w+\)return;//;
        $n_scroll++;
        "if($disabled)return;"
          . "if($key.meta&&\"j\"===$input)return void $scroll(1);"
          . "if($key.meta&&\"k\"===$input)return void $scroll(-1);"
          . "if($key.meta&&\"d\"===$input)return void $scroll(Math.max(1,$height-1));"
          . "if($key.meta&&\"u\"===$input)return void $scroll(-Math.max(1,$height-1));"
          . $rest
    }eg;

    my $n_input = 0;
    $src =~ s{$input_re}{
        my ( $input, $key, $disabled ) = ( $1, $2, $3 );
        $n_input++;
        $&
          . "if($key.meta&&(\"j\"===$input||\"k\"===$input||\"u\"===$input||\"d\"===$input))return;"
    }eg;

    next unless $n_scroll || $n_input;

    open my $out, '>', $file or die "cannot write $file: $!";
    print $out $src;
    close $out;

    $scroll_n += $n_scroll;
    $input_n  += $n_input;
    print STDERR "patched $file (scroll=$n_scroll input=$n_input)\n";
}

die "cursor-agent alt-scroll patch: expected exactly 1 scroll substitution, got $scroll_n\n"
  unless $scroll_n == 1;
die "cursor-agent alt-scroll patch: expected exactly 1 input substitution, got $input_n\n"
  unless $input_n == 1;
