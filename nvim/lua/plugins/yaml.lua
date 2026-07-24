return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      servers = {
        yamlls = {
          -- LazyVim yaml extraのbefore_initを置換: SchemaStoreカタログのRKEスキーマが
          -- fileMatch(cluster.yml/cluster.yaml)でK8sマニフェスト等に誤適用され
          -- "Property apiVersion is not allowed"等の誤検知を出すため除外する。
          -- name指定のignoreオプションはカタログ側の改名・削除でassertエラーになり
          -- yamlls起動ごと壊しうるため、urlキー削除(不在ならno-op)のfail-soft方式にする
          before_init = function(_, new_config)
            local schemas = require("schemastore").yaml.schemas()
            schemas["https://raw.githubusercontent.com/dcermak/vscode-rke-cluster-config/main/schemas/cluster.yml.json"] =
              nil
            new_config.settings.yaml.schemas =
              vim.tbl_deep_extend("force", new_config.settings.yaml.schemas or {}, schemas)
          end,
        },
      },
    },
  },
}
