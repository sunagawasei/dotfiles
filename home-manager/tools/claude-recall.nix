{ claude-recall, ... }:
let
  plugin = "${claude-recall}/share/claude-plugin/claude-recall";
in
{
  # 過去のClaude Codeセッションを SQLite へ溜めて検索する(recall CLI / MCP / TUI)
  home.packages = [ claude-recall ];

  home.file = {
    # ~/.claude/skills 直下のpluginディレクトリは claude-recall@skills-dir として読まれる。
    # バイナリと同じ版のpluginが入り、marketplace・`claude mcp add`・settings.jsonのhookが要らない
    ".claude/skills/claude-recall".source = plugin;
    # Codexには recall skill だけを渡す(plugin全体はClaude Code用)
    ".agents/skills/recall".source = "${plugin}/skills/recall";
  };
}
