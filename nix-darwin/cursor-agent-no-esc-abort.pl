#!/usr/bin/env perl
# cursor-agent の Esc/Ctrl+G による実行中断を落とす。CLI に設定が無いためバンドルを直接書き換える
use strict;
use warnings;

# queued message 取消の分岐(残す)と直後の abort 分岐(消す)をペアで縛る。
# 同じ escape 述語変数・同じ入力欄変数を後方参照で要求し、minify の変数名変更に耐えさせる
my $re =
qr/(if\((\w+)&&0===(\w+)\.length&&\w+&&(\w+)\)return void \4\(\);)if\(\2&&0===\3\.length&&\w+\)return null==(\w+)\|\|\5\(\),void\(0,\w+\.\w+\)\(\);/;

my $total = 0;
for my $file (@ARGV) {
    open my $in, '<', $file or die "cannot read $file: $!";
    my $src = do { local $/; <$in> };
    close $in;

    my $n = ( $src =~ s/$re/$1/g );
    next unless $n;

    open my $out, '>', $file or die "cannot write $file: $!";
    print $out $src;
    close $out;

    $total += $n;
    print STDERR "patched $file ($n)\n";
}

die "cursor-agent esc-abort patch: expected exactly 1 substitution, got $total\n"
  unless $total == 1;
