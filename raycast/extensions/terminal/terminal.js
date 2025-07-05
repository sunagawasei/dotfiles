"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/terminal.tsx
var terminal_exports = {};
__export(terminal_exports, {
  default: () => Command
});
module.exports = __toCommonJS(terminal_exports);
var import_api = require("@raycast/api");
var import_child_process = require("child_process");
var import_os = require("os");
var import_react = require("react");
var import_util = require("util");
var import_jsx_runtime = require("react/jsx-runtime");
var execAsync = (0, import_util.promisify)(import_child_process.exec);
var DEFAULT_COMMANDS = [
  {
    id: "1",
    name: "\u30DB\u30FC\u30E0\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3092\u958B\u304F",
    command: "",
    description: "\u30DB\u30FC\u30E0\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3067\u30BF\u30FC\u30DF\u30CA\u30EB\u3092\u958B\u304F",
    directory: (0, import_os.homedir)()
  },
  {
    id: "2",
    name: "\u73FE\u5728\u306E\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3092\u958B\u304F",
    command: "",
    description: "\u73FE\u5728\u306E\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3067\u30BF\u30FC\u30DF\u30CA\u30EB\u3092\u958B\u304F"
  },
  {
    id: "3",
    name: "Git Status",
    command: "git status",
    description: "Git\u306E\u72B6\u614B\u3092\u78BA\u8A8D"
  },
  {
    id: "4",
    name: "NPM Install",
    command: "npm install",
    description: "NPM\u30D1\u30C3\u30B1\u30FC\u30B8\u3092\u30A4\u30F3\u30B9\u30C8\u30FC\u30EB"
  },
  {
    id: "5",
    name: "List Files",
    command: "ls -la",
    description: "\u30D5\u30A1\u30A4\u30EB\u4E00\u89A7\u3092\u8868\u793A"
  }
];
function Command() {
  const [commands, setCommands] = (0, import_react.useState)([]);
  const [history, setHistory] = (0, import_react.useState)([]);
  const [isLoading, setIsLoading] = (0, import_react.useState)(true);
  const { push } = (0, import_api.useNavigation)();
  (0, import_react.useEffect)(() => {
    loadCommands();
    loadHistory();
  }, []);
  const loadCommands = async () => {
    try {
      const savedCommands = await import_api.LocalStorage.getItem("saved-commands");
      if (savedCommands) {
        setCommands(JSON.parse(savedCommands));
      } else {
        setCommands(DEFAULT_COMMANDS);
        await import_api.LocalStorage.setItem("saved-commands", JSON.stringify(DEFAULT_COMMANDS));
      }
    } catch (error) {
      console.error("Failed to load commands:", error);
      setCommands(DEFAULT_COMMANDS);
    } finally {
      setIsLoading(false);
    }
  };
  const loadHistory = async () => {
    try {
      const savedHistory = await import_api.LocalStorage.getItem("command-history");
      if (savedHistory) {
        const parsedHistory = JSON.parse(savedHistory);
        setHistory(parsedHistory.slice(-20));
      }
    } catch (error) {
      console.error("Failed to load history:", error);
    }
  };
  const saveCommands = async (newCommands) => {
    try {
      await import_api.LocalStorage.setItem("saved-commands", JSON.stringify(newCommands));
      setCommands(newCommands);
    } catch (error) {
      console.error("Failed to save commands:", error);
      (0, import_api.showToast)({
        style: import_api.Toast.Style.Failure,
        title: "Failed to save commands"
      });
    }
  };
  const saveToHistory = async (command, directory) => {
    try {
      const newHistoryItem = {
        id: Date.now().toString(),
        command,
        timestamp: Date.now(),
        directory
      };
      const updatedHistory = [...history, newHistoryItem].slice(-20);
      await import_api.LocalStorage.setItem("command-history", JSON.stringify(updatedHistory));
      setHistory(updatedHistory);
    } catch (error) {
      console.error("Failed to save to history:", error);
    }
  };
  const executeCommand = async (command) => {
    const preferences = (0, import_api.getPreferenceValues)();
    const targetDirectory = command.directory || preferences.defaultDirectory || (0, import_os.homedir)();
    if (!command.command) {
      push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(InteractiveTerminal, { directory: targetDirectory }));
    } else {
      push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(CommandExecutor, { command, directory: targetDirectory, onSaveHistory: saveToHistory }));
    }
  };
  const deleteCommand = async (commandId) => {
    const confirmed = await (0, import_api.confirmAlert)({
      title: "\u30B3\u30DE\u30F3\u30C9\u3092\u524A\u9664",
      message: "\u3053\u306E\u30B3\u30DE\u30F3\u30C9\u3092\u524A\u9664\u3057\u307E\u3059\u304B\uFF1F",
      primaryAction: {
        title: "\u524A\u9664",
        style: import_api.Alert.ActionStyle.Destructive
      }
    });
    if (confirmed) {
      const updatedCommands = commands.filter((cmd) => cmd.id !== commandId);
      await saveCommands(updatedCommands);
      (0, import_api.showToast)({
        style: import_api.Toast.Style.Success,
        title: "\u30B3\u30DE\u30F3\u30C9\u3092\u524A\u9664\u3057\u307E\u3057\u305F"
      });
    }
  };
  const clearHistory = async () => {
    const confirmed = await (0, import_api.confirmAlert)({
      title: "\u5C65\u6B74\u3092\u30AF\u30EA\u30A2",
      message: "\u30B3\u30DE\u30F3\u30C9\u5C65\u6B74\u3092\u3059\u3079\u3066\u524A\u9664\u3057\u307E\u3059\u304B\uFF1F",
      primaryAction: {
        title: "\u524A\u9664",
        style: import_api.Alert.ActionStyle.Destructive
      }
    });
    if (confirmed) {
      await import_api.LocalStorage.removeItem("command-history");
      setHistory([]);
      (0, import_api.showToast)({
        style: import_api.Toast.Style.Success,
        title: "\u5C65\u6B74\u3092\u30AF\u30EA\u30A2\u3057\u307E\u3057\u305F"
      });
    }
  };
  return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.List, { isLoading, searchBarPlaceholder: "\u30B3\u30DE\u30F3\u30C9\u3092\u691C\u7D22...", children: [
    /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.List.Section, { title: "\u3088\u304F\u4F7F\u3046\u30B3\u30DE\u30F3\u30C9", children: commands.map((command) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
      import_api.List.Item,
      {
        title: command.name,
        subtitle: command.command || "\u30BF\u30FC\u30DF\u30CA\u30EB\u3092\u958B\u304F",
        accessories: [{ text: command.directory ? `\u{1F4C1} ${command.directory}` : "" }],
        actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u5B9F\u884C", icon: import_api.Icon.Terminal, onAction: () => executeCommand(command) }),
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u7DE8\u96C6",
              icon: import_api.Icon.Pencil,
              onAction: () => push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(EditCommandForm, { command, onSave: saveCommands, commands }))
            }
          ),
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u524A\u9664",
              icon: import_api.Icon.Trash,
              style: import_api.Action.Style.Destructive,
              onAction: () => deleteCommand(command.id)
            }
          ),
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u65B0\u3057\u3044\u30B3\u30DE\u30F3\u30C9\u3092\u8FFD\u52A0",
              icon: import_api.Icon.Plus,
              onAction: () => push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(AddCommandForm, { onSave: saveCommands, commands })),
              shortcut: { modifiers: ["cmd"], key: "n" }
            }
          ),
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u5C65\u6B74\u3092\u8868\u793A",
              icon: import_api.Icon.Clock,
              onAction: () => push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(HistoryView, { history, onClear: clearHistory }))
            }
          )
        ] })
      },
      command.id
    )) }),
    history.length > 0 && /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.List.Section, { title: "\u6700\u8FD1\u306E\u5C65\u6B74", children: history.slice(-5).reverse().map((item) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
      import_api.List.Item,
      {
        title: item.command,
        subtitle: `${item.directory} \u2022 ${new Date(item.timestamp).toLocaleString()}`,
        icon: import_api.Icon.Clock,
        actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u518D\u5B9F\u884C",
              icon: import_api.Icon.Terminal,
              onAction: () => {
                const cmd = {
                  id: item.id,
                  name: item.command,
                  command: item.command,
                  directory: item.directory
                };
                push(
                  /* @__PURE__ */ (0, import_jsx_runtime.jsx)(CommandExecutor, { command: cmd, directory: item.directory, onSaveHistory: saveToHistory })
                );
              }
            }
          ),
          /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
            import_api.Action,
            {
              title: "\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3067\u958B\u304F",
              icon: import_api.Icon.Folder,
              onAction: () => push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(InteractiveTerminal, { directory: item.directory }))
            }
          )
        ] })
      },
      item.id
    )) })
  ] });
}
function AddCommandForm({ onSave, commands }) {
  const { pop } = (0, import_api.useNavigation)();
  const [nameError, setNameError] = (0, import_react.useState)();
  const [commandError, setCommandError] = (0, import_react.useState)();
  function handleSubmit(values) {
    if (!values.name.trim()) {
      setNameError("\u540D\u524D\u306F\u5FC5\u9808\u3067\u3059");
      return;
    }
    const newCommand = {
      id: Date.now().toString(),
      name: values.name.trim(),
      command: values.command.trim(),
      description: values.description.trim(),
      directory: values.directory.trim() || void 0
    };
    const updatedCommands = [...commands, newCommand];
    onSave(updatedCommands);
    (0, import_api.showToast)({
      style: import_api.Toast.Style.Success,
      title: "\u30B3\u30DE\u30F3\u30C9\u3092\u8FFD\u52A0\u3057\u307E\u3057\u305F"
    });
    pop();
  }
  return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(
    import_api.Form,
    {
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action.SubmitForm, { title: "\u8FFD\u52A0", onSubmit: handleSubmit }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u30AD\u30E3\u30F3\u30BB\u30EB", onAction: pop })
      ] }),
      children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Form.TextField,
          {
            id: "name",
            title: "\u540D\u524D",
            placeholder: "\u30B3\u30DE\u30F3\u30C9\u306E\u540D\u524D\u3092\u5165\u529B",
            error: nameError,
            onChange: () => setNameError(void 0)
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Form.TextField,
          {
            id: "command",
            title: "\u30B3\u30DE\u30F3\u30C9",
            placeholder: "\u5B9F\u884C\u3059\u308B\u30B3\u30DE\u30F3\u30C9\u3092\u5165\u529B\uFF08\u7A7A\u306E\u5834\u5408\u306F\u30BF\u30FC\u30DF\u30CA\u30EB\u306E\u307F\u958B\u304F\uFF09",
            error: commandError,
            onChange: () => setCommandError(void 0)
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.TextField, { id: "description", title: "\u8AAC\u660E", placeholder: "\u30B3\u30DE\u30F3\u30C9\u306E\u8AAC\u660E\uFF08\u30AA\u30D7\u30B7\u30E7\u30F3\uFF09" }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.TextField, { id: "directory", title: "\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA", placeholder: "\u5B9F\u884C\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\uFF08\u7A7A\u306E\u5834\u5408\u306F\u30C7\u30D5\u30A9\u30EB\u30C8\uFF09" })
      ]
    }
  );
}
function EditCommandForm({
  command,
  onSave,
  commands
}) {
  const { pop } = (0, import_api.useNavigation)();
  const [nameError, setNameError] = (0, import_react.useState)();
  function handleSubmit(values) {
    if (!values.name.trim()) {
      setNameError("\u540D\u524D\u306F\u5FC5\u9808\u3067\u3059");
      return;
    }
    const updatedCommand = {
      ...command,
      name: values.name.trim(),
      command: values.command.trim(),
      description: values.description.trim(),
      directory: values.directory.trim() || void 0
    };
    const updatedCommands = commands.map((cmd) => cmd.id === command.id ? updatedCommand : cmd);
    onSave(updatedCommands);
    (0, import_api.showToast)({
      style: import_api.Toast.Style.Success,
      title: "\u30B3\u30DE\u30F3\u30C9\u3092\u66F4\u65B0\u3057\u307E\u3057\u305F"
    });
    pop();
  }
  return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(
    import_api.Form,
    {
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action.SubmitForm, { title: "\u4FDD\u5B58", onSubmit: handleSubmit }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u30AD\u30E3\u30F3\u30BB\u30EB", onAction: pop })
      ] }),
      children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Form.TextField,
          {
            id: "name",
            title: "\u540D\u524D",
            defaultValue: command.name,
            error: nameError,
            onChange: () => setNameError(void 0)
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.TextField, { id: "command", title: "\u30B3\u30DE\u30F3\u30C9", defaultValue: command.command }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.TextField, { id: "description", title: "\u8AAC\u660E", defaultValue: command.description || "" }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.TextField, { id: "directory", title: "\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA", defaultValue: command.directory || "" })
      ]
    }
  );
}
function HistoryView({ history, onClear }) {
  const { push } = (0, import_api.useNavigation)();
  return /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.List, { searchBarPlaceholder: "\u5C65\u6B74\u3092\u691C\u7D22...", children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.List.Section, { title: `\u30B3\u30DE\u30F3\u30C9\u5C65\u6B74 (${history.length}\u4EF6)`, children: history.slice().reverse().map((item) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
    import_api.List.Item,
    {
      title: item.command,
      subtitle: item.directory,
      accessories: [{ text: new Date(item.timestamp).toLocaleString() }],
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action,
          {
            title: "\u518D\u5B9F\u884C",
            icon: import_api.Icon.Terminal,
            onAction: () => {
              const cmd = {
                id: item.id,
                name: item.command,
                command: item.command,
                directory: item.directory
              };
              push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(CommandExecutor, { command: cmd, directory: item.directory }));
            }
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action,
          {
            title: "\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u3067\u958B\u304F",
            icon: import_api.Icon.Folder,
            onAction: () => push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(InteractiveTerminal, { directory: item.directory }))
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u5C65\u6B74\u3092\u30AF\u30EA\u30A2", icon: import_api.Icon.Trash, style: import_api.Action.Style.Destructive, onAction: onClear })
      ] })
    },
    item.id
  )) }) });
}
function CommandExecutor({
  command,
  directory,
  onSaveHistory
}) {
  const [result, setResult] = (0, import_react.useState)(null);
  const [isLoading, setIsLoading] = (0, import_react.useState)(true);
  const { pop } = (0, import_api.useNavigation)();
  (0, import_react.useEffect)(() => {
    executeCommand();
  }, []);
  const executeCommand = async () => {
    const startTime = Date.now();
    try {
      const { stdout, stderr } = await execAsync(command.command, {
        cwd: directory,
        env: { ...process.env, PWD: directory },
        maxBuffer: 1024 * 1024 * 10
        // 10MB
      });
      const endTime = Date.now();
      const commandResult = {
        id: Date.now().toString(),
        command: command.command,
        directory,
        output: stdout,
        error: stderr,
        exitCode: 0,
        timestamp: startTime,
        duration: endTime - startTime
      };
      setResult(commandResult);
      if (onSaveHistory) {
        await onSaveHistory(command.command, directory);
      }
    } catch (error) {
      const endTime = Date.now();
      const errorObj = error;
      const commandResult = {
        id: Date.now().toString(),
        command: command.command,
        directory,
        output: errorObj.stdout || "",
        error: errorObj.stderr || errorObj.message || "Unknown error",
        exitCode: errorObj.code || 1,
        timestamp: startTime,
        duration: endTime - startTime
      };
      setResult(commandResult);
      if (onSaveHistory) {
        await onSaveHistory(command.command, directory);
      }
    } finally {
      setIsLoading(false);
    }
  };
  const getMarkdown = () => {
    if (!result) return "";
    let markdown = `# \u30B3\u30DE\u30F3\u30C9\u5B9F\u884C\u7D50\u679C

**\u30B3\u30DE\u30F3\u30C9:** \`${result.command}\`  
**\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA:** \`${result.directory}\`  
**\u5B9F\u884C\u6642\u9593:** ${result.duration}ms  
**\u7D42\u4E86\u30B3\u30FC\u30C9:** ${result.exitCode}  
**\u5B9F\u884C\u65E5\u6642:** ${new Date(result.timestamp).toLocaleString()}

`;
    if (result.output) {
      markdown += `## \u51FA\u529B

\`\`\`
${result.output}
\`\`\`

`;
    }
    if (result.error) {
      markdown += `## \u30A8\u30E9\u30FC

\`\`\`
${result.error}
\`\`\`

`;
    }
    return markdown;
  };
  return /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
    import_api.Detail,
    {
      isLoading,
      markdown: getMarkdown(),
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action,
          {
            title: "\u518D\u5B9F\u884C",
            icon: import_api.Icon.ArrowClockwise,
            onAction: () => {
              setIsLoading(true);
              executeCommand();
            }
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u623B\u308B", icon: import_api.Icon.ArrowLeft, onAction: pop, shortcut: { modifiers: ["cmd"], key: "w" } }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action.CopyToClipboard,
          {
            title: "\u51FA\u529B\u3092\u30B3\u30D4\u30FC",
            content: result?.output || "",
            shortcut: { modifiers: ["cmd"], key: "c" }
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action.CopyToClipboard,
          {
            title: "\u30B3\u30DE\u30F3\u30C9\u3092\u30B3\u30D4\u30FC",
            content: result?.command || "",
            shortcut: { modifiers: ["cmd", "shift"], key: "c" }
          }
        )
      ] })
    }
  );
}
function InteractiveTerminal({ directory }) {
  const [commandHistory, setCommandHistory] = (0, import_react.useState)([]);
  const [results, setResults] = (0, import_react.useState)([]);
  const { pop, push } = (0, import_api.useNavigation)();
  const executeCommand = async (command) => {
    if (!command.trim()) return;
    const startTime = Date.now();
    try {
      const { stdout, stderr } = await execAsync(command, {
        cwd: directory,
        env: { ...process.env, PWD: directory },
        maxBuffer: 1024 * 1024 * 10
      });
      const endTime = Date.now();
      const result = {
        id: Date.now().toString(),
        command,
        directory,
        output: stdout,
        error: stderr,
        exitCode: 0,
        timestamp: startTime,
        duration: endTime - startTime
      };
      setResults((prev) => [...prev, result]);
      setCommandHistory((prev) => [...prev, command]);
    } catch (error) {
      const endTime = Date.now();
      const errorObj = error;
      const result = {
        id: Date.now().toString(),
        command,
        directory,
        output: errorObj.stdout || "",
        error: errorObj.stderr || errorObj.message || "Unknown error",
        exitCode: errorObj.code || 1,
        timestamp: startTime,
        duration: endTime - startTime
      };
      setResults((prev) => [...prev, result]);
      setCommandHistory((prev) => [...prev, command]);
    }
  };
  const getTerminalMarkdown = () => {
    let markdown = `# \u30A4\u30F3\u30BF\u30E9\u30AF\u30C6\u30A3\u30D6\u30BF\u30FC\u30DF\u30CA\u30EB

**\u73FE\u5728\u306E\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA:** \`${directory}\`

---

`;
    results.forEach((result) => {
      markdown += `## \`${result.command}\`
**\u5B9F\u884C\u6642\u9593:** ${result.duration}ms | **\u7D42\u4E86\u30B3\u30FC\u30C9:** ${result.exitCode} | **\u5B9F\u884C\u6642\u523B:** ${new Date(result.timestamp).toLocaleTimeString()}

`;
      if (result.output) {
        markdown += `\`\`\`
${result.output}
\`\`\`

`;
      }
      if (result.error) {
        markdown += `**\u30A8\u30E9\u30FC:**
\`\`\`
${result.error}
\`\`\`

`;
      }
      markdown += "---\n\n";
    });
    if (results.length === 0) {
      markdown += `\u30B3\u30DE\u30F3\u30C9\u3092\u5165\u529B\u3057\u3066\u5B9F\u884C\u3057\u3066\u304F\u3060\u3055\u3044\u3002

**\u4F7F\u7528\u4F8B:**
- \`ls -la\` - \u30D5\u30A1\u30A4\u30EB\u4E00\u89A7\u8868\u793A
- \`pwd\` - \u73FE\u5728\u306E\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA\u8868\u793A
- \`git status\` - Git\u72B6\u614B\u78BA\u8A8D
- \`npm install\` - \u30D1\u30C3\u30B1\u30FC\u30B8\u30A4\u30F3\u30B9\u30C8\u30FC\u30EB

`;
    }
    return markdown;
  };
  return /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
    import_api.Detail,
    {
      markdown: getTerminalMarkdown(),
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action,
          {
            title: "\u30B3\u30DE\u30F3\u30C9\u3092\u5165\u529B",
            icon: import_api.Icon.Terminal,
            onAction: () => {
              push(/* @__PURE__ */ (0, import_jsx_runtime.jsx)(CommandInputForm, { onExecute: executeCommand, directory }));
            }
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u623B\u308B", icon: import_api.Icon.ArrowLeft, onAction: pop, shortcut: { modifiers: ["cmd"], key: "w" } }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Action,
          {
            title: "\u5C65\u6B74\u3092\u30AF\u30EA\u30A2",
            icon: import_api.Icon.Trash,
            style: import_api.Action.Style.Destructive,
            onAction: () => {
              setResults([]);
              setCommandHistory([]);
            }
          }
        )
      ] }),
      metadata: /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Detail.Metadata, { children: /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Detail.Metadata.TagList, { title: "\u5B9F\u884C\u6E08\u307F\u30B3\u30DE\u30F3\u30C9", children: commandHistory.slice(-5).map((cmd, index) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Detail.Metadata.TagList.Item, { text: cmd }, index)) }) })
    }
  );
}
function CommandInputForm({ onExecute, directory }) {
  const { pop } = (0, import_api.useNavigation)();
  const [commandError, setCommandError] = (0, import_react.useState)();
  const commonCommands = [
    "ls -la",
    "pwd",
    "git status",
    "git log --oneline",
    "npm install",
    "npm run build",
    "npm test",
    "docker ps",
    "ps aux",
    "df -h"
  ];
  function handleSubmit(values) {
    if (!values.command.trim()) {
      setCommandError("\u30B3\u30DE\u30F3\u30C9\u3092\u5165\u529B\u3057\u3066\u304F\u3060\u3055\u3044");
      return;
    }
    onExecute(values.command.trim());
    pop();
  }
  return /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(
    import_api.Form,
    {
      actions: /* @__PURE__ */ (0, import_jsx_runtime.jsxs)(import_api.ActionPanel, { children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action.SubmitForm, { title: "\u5B9F\u884C", onSubmit: handleSubmit }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Action, { title: "\u30AD\u30E3\u30F3\u30BB\u30EB", onAction: pop })
      ] }),
      children: [
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.Description, { text: `\u5B9F\u884C\u30C7\u30A3\u30EC\u30AF\u30C8\u30EA: ${directory}` }),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(
          import_api.Form.TextField,
          {
            id: "command",
            title: "\u30B3\u30DE\u30F3\u30C9",
            placeholder: "\u5B9F\u884C\u3059\u308B\u30B3\u30DE\u30F3\u30C9\u3092\u5165\u529B\u3057\u3066\u304F\u3060\u3055\u3044",
            error: commandError,
            onChange: () => setCommandError(void 0)
          }
        ),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.Separator, {}),
        /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.Description, { title: "\u3088\u304F\u4F7F\u3046\u30B3\u30DE\u30F3\u30C9", text: "\u4EE5\u4E0B\u306E\u30B3\u30DE\u30F3\u30C9\u3092\u53C2\u8003\u306B\u3057\u3066\u304F\u3060\u3055\u3044:" }),
        commonCommands.map((cmd, index) => /* @__PURE__ */ (0, import_jsx_runtime.jsx)(import_api.Form.Description, { text: `\u2022 ${cmd}` }, index))
      ]
    }
  );
}
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsiLi4vLi4vLi4vLi4vcG9jL3Rlcm1pbmFsL3NyYy90ZXJtaW5hbC50c3giXSwKICAic291cmNlc0NvbnRlbnQiOiBbImltcG9ydCB7XG4gIEFjdGlvbixcbiAgQWN0aW9uUGFuZWwsXG4gIEFsZXJ0LFxuICBjb25maXJtQWxlcnQsXG4gIERldGFpbCxcbiAgRm9ybSxcbiAgZ2V0UHJlZmVyZW5jZVZhbHVlcyxcbiAgSWNvbixcbiAgTGlzdCxcbiAgTG9jYWxTdG9yYWdlLFxuICBzaG93VG9hc3QsXG4gIFRvYXN0LFxuICB1c2VOYXZpZ2F0aW9uLFxufSBmcm9tIFwiQHJheWNhc3QvYXBpXCI7XG5pbXBvcnQgeyBleGVjIH0gZnJvbSBcImNoaWxkX3Byb2Nlc3NcIjtcbmltcG9ydCB7IGhvbWVkaXIgfSBmcm9tIFwib3NcIjtcbmltcG9ydCB7IHVzZUVmZmVjdCwgdXNlU3RhdGUgfSBmcm9tIFwicmVhY3RcIjtcbmltcG9ydCB7IHByb21pc2lmeSB9IGZyb20gXCJ1dGlsXCI7XG5cbmNvbnN0IGV4ZWNBc3luYyA9IHByb21pc2lmeShleGVjKTtcblxuaW50ZXJmYWNlIFByZWZlcmVuY2VzIHtcbiAgdGVybWluYWxBcHA6IHN0cmluZztcbiAgZGVmYXVsdERpcmVjdG9yeTogc3RyaW5nO1xufVxuXG5pbnRlcmZhY2UgQ29tbWFuZCB7XG4gIGlkOiBzdHJpbmc7XG4gIG5hbWU6IHN0cmluZztcbiAgY29tbWFuZDogc3RyaW5nO1xuICBkZXNjcmlwdGlvbj86IHN0cmluZztcbiAgZGlyZWN0b3J5Pzogc3RyaW5nO1xufVxuXG5pbnRlcmZhY2UgQ29tbWFuZFJlc3VsdCB7XG4gIGlkOiBzdHJpbmc7XG4gIGNvbW1hbmQ6IHN0cmluZztcbiAgZGlyZWN0b3J5OiBzdHJpbmc7XG4gIG91dHB1dDogc3RyaW5nO1xuICBlcnJvcj86IHN0cmluZztcbiAgZXhpdENvZGU/OiBudW1iZXI7XG4gIHRpbWVzdGFtcDogbnVtYmVyO1xuICBkdXJhdGlvbjogbnVtYmVyO1xufVxuXG5pbnRlcmZhY2UgQ29tbWFuZEhpc3Rvcnkge1xuICBpZDogc3RyaW5nO1xuICBjb21tYW5kOiBzdHJpbmc7XG4gIHRpbWVzdGFtcDogbnVtYmVyO1xuICBkaXJlY3Rvcnk6IHN0cmluZztcbn1cblxuLy8gXHUzMEM3XHUzMEQ1XHUzMEE5XHUzMEVCXHUzMEM4XHUzMDZFXHUzMDg4XHUzMDRGXHU0RjdGXHUzMDQ2XHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XG5jb25zdCBERUZBVUxUX0NPTU1BTkRTOiBDb21tYW5kW10gPSBbXG4gIHtcbiAgICBpZDogXCIxXCIsXG4gICAgbmFtZTogXCJcdTMwREJcdTMwRkNcdTMwRTBcdTMwQzdcdTMwQTNcdTMwRUNcdTMwQUZcdTMwQzhcdTMwRUFcdTMwOTJcdTk1OEJcdTMwNEZcIixcbiAgICBjb21tYW5kOiBcIlwiLFxuICAgIGRlc2NyaXB0aW9uOiBcIlx1MzBEQlx1MzBGQ1x1MzBFMFx1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQVx1MzA2N1x1MzBCRlx1MzBGQ1x1MzBERlx1MzBDQVx1MzBFQlx1MzA5Mlx1OTU4Qlx1MzA0RlwiLFxuICAgIGRpcmVjdG9yeTogaG9tZWRpcigpLFxuICB9LFxuICB7XG4gICAgaWQ6IFwiMlwiLFxuICAgIG5hbWU6IFwiXHU3M0ZFXHU1NzI4XHUzMDZFXHUzMEM3XHUzMEEzXHUzMEVDXHUzMEFGXHUzMEM4XHUzMEVBXHUzMDkyXHU5NThCXHUzMDRGXCIsXG4gICAgY29tbWFuZDogXCJcIixcbiAgICBkZXNjcmlwdGlvbjogXCJcdTczRkVcdTU3MjhcdTMwNkVcdTMwQzdcdTMwQTNcdTMwRUNcdTMwQUZcdTMwQzhcdTMwRUFcdTMwNjdcdTMwQkZcdTMwRkNcdTMwREZcdTMwQ0FcdTMwRUJcdTMwOTJcdTk1OEJcdTMwNEZcIixcbiAgfSxcbiAge1xuICAgIGlkOiBcIjNcIixcbiAgICBuYW1lOiBcIkdpdCBTdGF0dXNcIixcbiAgICBjb21tYW5kOiBcImdpdCBzdGF0dXNcIixcbiAgICBkZXNjcmlwdGlvbjogXCJHaXRcdTMwNkVcdTcyQjZcdTYxNEJcdTMwOTJcdTc4QkFcdThBOERcIixcbiAgfSxcbiAge1xuICAgIGlkOiBcIjRcIixcbiAgICBuYW1lOiBcIk5QTSBJbnN0YWxsXCIsXG4gICAgY29tbWFuZDogXCJucG0gaW5zdGFsbFwiLFxuICAgIGRlc2NyaXB0aW9uOiBcIk5QTVx1MzBEMVx1MzBDM1x1MzBCMVx1MzBGQ1x1MzBCOFx1MzA5Mlx1MzBBNFx1MzBGM1x1MzBCOVx1MzBDOFx1MzBGQ1x1MzBFQlwiLFxuICB9LFxuICB7XG4gICAgaWQ6IFwiNVwiLFxuICAgIG5hbWU6IFwiTGlzdCBGaWxlc1wiLFxuICAgIGNvbW1hbmQ6IFwibHMgLWxhXCIsXG4gICAgZGVzY3JpcHRpb246IFwiXHUzMEQ1XHUzMEExXHUzMEE0XHUzMEVCXHU0RTAwXHU4OUE3XHUzMDkyXHU4ODY4XHU3OTNBXCIsXG4gIH0sXG5dO1xuXG5leHBvcnQgZGVmYXVsdCBmdW5jdGlvbiBDb21tYW5kKCkge1xuICBjb25zdCBbY29tbWFuZHMsIHNldENvbW1hbmRzXSA9IHVzZVN0YXRlPENvbW1hbmRbXT4oW10pO1xuICBjb25zdCBbaGlzdG9yeSwgc2V0SGlzdG9yeV0gPSB1c2VTdGF0ZTxDb21tYW5kSGlzdG9yeVtdPihbXSk7XG4gIGNvbnN0IFtpc0xvYWRpbmcsIHNldElzTG9hZGluZ10gPSB1c2VTdGF0ZSh0cnVlKTtcbiAgY29uc3QgeyBwdXNoIH0gPSB1c2VOYXZpZ2F0aW9uKCk7XG5cbiAgdXNlRWZmZWN0KCgpID0+IHtcbiAgICBsb2FkQ29tbWFuZHMoKTtcbiAgICBsb2FkSGlzdG9yeSgpO1xuICB9LCBbXSk7XG5cbiAgY29uc3QgbG9hZENvbW1hbmRzID0gYXN5bmMgKCkgPT4ge1xuICAgIHRyeSB7XG4gICAgICBjb25zdCBzYXZlZENvbW1hbmRzID0gYXdhaXQgTG9jYWxTdG9yYWdlLmdldEl0ZW08c3RyaW5nPihcInNhdmVkLWNvbW1hbmRzXCIpO1xuICAgICAgaWYgKHNhdmVkQ29tbWFuZHMpIHtcbiAgICAgICAgc2V0Q29tbWFuZHMoSlNPTi5wYXJzZShzYXZlZENvbW1hbmRzKSk7XG4gICAgICB9IGVsc2Uge1xuICAgICAgICBzZXRDb21tYW5kcyhERUZBVUxUX0NPTU1BTkRTKTtcbiAgICAgICAgYXdhaXQgTG9jYWxTdG9yYWdlLnNldEl0ZW0oXCJzYXZlZC1jb21tYW5kc1wiLCBKU09OLnN0cmluZ2lmeShERUZBVUxUX0NPTU1BTkRTKSk7XG4gICAgICB9XG4gICAgfSBjYXRjaCAoZXJyb3IpIHtcbiAgICAgIGNvbnNvbGUuZXJyb3IoXCJGYWlsZWQgdG8gbG9hZCBjb21tYW5kczpcIiwgZXJyb3IpO1xuICAgICAgc2V0Q29tbWFuZHMoREVGQVVMVF9DT01NQU5EUyk7XG4gICAgfSBmaW5hbGx5IHtcbiAgICAgIHNldElzTG9hZGluZyhmYWxzZSk7XG4gICAgfVxuICB9O1xuXG4gIGNvbnN0IGxvYWRIaXN0b3J5ID0gYXN5bmMgKCkgPT4ge1xuICAgIHRyeSB7XG4gICAgICBjb25zdCBzYXZlZEhpc3RvcnkgPSBhd2FpdCBMb2NhbFN0b3JhZ2UuZ2V0SXRlbTxzdHJpbmc+KFwiY29tbWFuZC1oaXN0b3J5XCIpO1xuICAgICAgaWYgKHNhdmVkSGlzdG9yeSkge1xuICAgICAgICBjb25zdCBwYXJzZWRIaXN0b3J5ID0gSlNPTi5wYXJzZShzYXZlZEhpc3RvcnkpO1xuICAgICAgICBzZXRIaXN0b3J5KHBhcnNlZEhpc3Rvcnkuc2xpY2UoLTIwKSk7IC8vIFx1NjcwMFx1NjVCMDIwXHU0RUY2XHUzMDZFXHUzMDdGXHU0RkREXHU2MzAxXG4gICAgICB9XG4gICAgfSBjYXRjaCAoZXJyb3IpIHtcbiAgICAgIGNvbnNvbGUuZXJyb3IoXCJGYWlsZWQgdG8gbG9hZCBoaXN0b3J5OlwiLCBlcnJvcik7XG4gICAgfVxuICB9O1xuXG4gIGNvbnN0IHNhdmVDb21tYW5kcyA9IGFzeW5jIChuZXdDb21tYW5kczogQ29tbWFuZFtdKSA9PiB7XG4gICAgdHJ5IHtcbiAgICAgIGF3YWl0IExvY2FsU3RvcmFnZS5zZXRJdGVtKFwic2F2ZWQtY29tbWFuZHNcIiwgSlNPTi5zdHJpbmdpZnkobmV3Q29tbWFuZHMpKTtcbiAgICAgIHNldENvbW1hbmRzKG5ld0NvbW1hbmRzKTtcbiAgICB9IGNhdGNoIChlcnJvcikge1xuICAgICAgY29uc29sZS5lcnJvcihcIkZhaWxlZCB0byBzYXZlIGNvbW1hbmRzOlwiLCBlcnJvcik7XG4gICAgICBzaG93VG9hc3Qoe1xuICAgICAgICBzdHlsZTogVG9hc3QuU3R5bGUuRmFpbHVyZSxcbiAgICAgICAgdGl0bGU6IFwiRmFpbGVkIHRvIHNhdmUgY29tbWFuZHNcIixcbiAgICAgIH0pO1xuICAgIH1cbiAgfTtcblxuICBjb25zdCBzYXZlVG9IaXN0b3J5ID0gYXN5bmMgKGNvbW1hbmQ6IHN0cmluZywgZGlyZWN0b3J5OiBzdHJpbmcpID0+IHtcbiAgICB0cnkge1xuICAgICAgY29uc3QgbmV3SGlzdG9yeUl0ZW06IENvbW1hbmRIaXN0b3J5ID0ge1xuICAgICAgICBpZDogRGF0ZS5ub3coKS50b1N0cmluZygpLFxuICAgICAgICBjb21tYW5kLFxuICAgICAgICB0aW1lc3RhbXA6IERhdGUubm93KCksXG4gICAgICAgIGRpcmVjdG9yeSxcbiAgICAgIH07XG4gICAgICBjb25zdCB1cGRhdGVkSGlzdG9yeSA9IFsuLi5oaXN0b3J5LCBuZXdIaXN0b3J5SXRlbV0uc2xpY2UoLTIwKTtcbiAgICAgIGF3YWl0IExvY2FsU3RvcmFnZS5zZXRJdGVtKFwiY29tbWFuZC1oaXN0b3J5XCIsIEpTT04uc3RyaW5naWZ5KHVwZGF0ZWRIaXN0b3J5KSk7XG4gICAgICBzZXRIaXN0b3J5KHVwZGF0ZWRIaXN0b3J5KTtcbiAgICB9IGNhdGNoIChlcnJvcikge1xuICAgICAgY29uc29sZS5lcnJvcihcIkZhaWxlZCB0byBzYXZlIHRvIGhpc3Rvcnk6XCIsIGVycm9yKTtcbiAgICB9XG4gIH07XG5cbiAgY29uc3QgZXhlY3V0ZUNvbW1hbmQgPSBhc3luYyAoY29tbWFuZDogQ29tbWFuZCkgPT4ge1xuICAgIGNvbnN0IHByZWZlcmVuY2VzID0gZ2V0UHJlZmVyZW5jZVZhbHVlczxQcmVmZXJlbmNlcz4oKTtcbiAgICBjb25zdCB0YXJnZXREaXJlY3RvcnkgPSBjb21tYW5kLmRpcmVjdG9yeSB8fCBwcmVmZXJlbmNlcy5kZWZhdWx0RGlyZWN0b3J5IHx8IGhvbWVkaXIoKTtcblxuICAgIGlmICghY29tbWFuZC5jb21tYW5kKSB7XG4gICAgICAvLyBcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwNENcdTdBN0FcdTMwNkVcdTU4MzRcdTU0MDhcdTMwNkZcdTMwMDFcdTMwQTRcdTMwRjNcdTMwQkZcdTMwRTlcdTMwQUZcdTMwQzZcdTMwQTNcdTMwRDZcdTMwQkZcdTMwRkNcdTMwREZcdTMwQ0FcdTMwRUJcdTMwOTJcdTk1OEJcdTMwNEZcbiAgICAgIHB1c2goPEludGVyYWN0aXZlVGVybWluYWwgZGlyZWN0b3J5PXt0YXJnZXREaXJlY3Rvcnl9IC8+KTtcbiAgICB9IGVsc2Uge1xuICAgICAgLy8gXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XHUzMDkyXHU1QjlGXHU4ODRDXHUzMDU3XHUzMDY2XHU3RDUwXHU2NzlDXHUzMDkyXHU4ODY4XHU3OTNBXG4gICAgICBwdXNoKDxDb21tYW5kRXhlY3V0b3IgY29tbWFuZD17Y29tbWFuZH0gZGlyZWN0b3J5PXt0YXJnZXREaXJlY3Rvcnl9IG9uU2F2ZUhpc3Rvcnk9e3NhdmVUb0hpc3Rvcnl9IC8+KTtcbiAgICB9XG4gIH07XG5cbiAgY29uc3QgZGVsZXRlQ29tbWFuZCA9IGFzeW5jIChjb21tYW5kSWQ6IHN0cmluZykgPT4ge1xuICAgIGNvbnN0IGNvbmZpcm1lZCA9IGF3YWl0IGNvbmZpcm1BbGVydCh7XG4gICAgICB0aXRsZTogXCJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTUyNEFcdTk2NjRcIixcbiAgICAgIG1lc3NhZ2U6IFwiXHUzMDUzXHUzMDZFXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XHUzMDkyXHU1MjRBXHU5NjY0XHUzMDU3XHUzMDdFXHUzMDU5XHUzMDRCXHVGRjFGXCIsXG4gICAgICBwcmltYXJ5QWN0aW9uOiB7XG4gICAgICAgIHRpdGxlOiBcIlx1NTI0QVx1OTY2NFwiLFxuICAgICAgICBzdHlsZTogQWxlcnQuQWN0aW9uU3R5bGUuRGVzdHJ1Y3RpdmUsXG4gICAgICB9LFxuICAgIH0pO1xuXG4gICAgaWYgKGNvbmZpcm1lZCkge1xuICAgICAgY29uc3QgdXBkYXRlZENvbW1hbmRzID0gY29tbWFuZHMuZmlsdGVyKChjbWQpID0+IGNtZC5pZCAhPT0gY29tbWFuZElkKTtcbiAgICAgIGF3YWl0IHNhdmVDb21tYW5kcyh1cGRhdGVkQ29tbWFuZHMpO1xuICAgICAgc2hvd1RvYXN0KHtcbiAgICAgICAgc3R5bGU6IFRvYXN0LlN0eWxlLlN1Y2Nlc3MsXG4gICAgICAgIHRpdGxlOiBcIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA5Mlx1NTI0QVx1OTY2NFx1MzA1N1x1MzA3RVx1MzA1N1x1MzA1RlwiLFxuICAgICAgfSk7XG4gICAgfVxuICB9O1xuXG4gIGNvbnN0IGNsZWFySGlzdG9yeSA9IGFzeW5jICgpID0+IHtcbiAgICBjb25zdCBjb25maXJtZWQgPSBhd2FpdCBjb25maXJtQWxlcnQoe1xuICAgICAgdGl0bGU6IFwiXHU1QzY1XHU2Qjc0XHUzMDkyXHUzMEFGXHUzMEVBXHUzMEEyXCIsXG4gICAgICBtZXNzYWdlOiBcIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1NUM2NVx1NkI3NFx1MzA5Mlx1MzA1OVx1MzA3OVx1MzA2Nlx1NTI0QVx1OTY2NFx1MzA1N1x1MzA3RVx1MzA1OVx1MzA0Qlx1RkYxRlwiLFxuICAgICAgcHJpbWFyeUFjdGlvbjoge1xuICAgICAgICB0aXRsZTogXCJcdTUyNEFcdTk2NjRcIixcbiAgICAgICAgc3R5bGU6IEFsZXJ0LkFjdGlvblN0eWxlLkRlc3RydWN0aXZlLFxuICAgICAgfSxcbiAgICB9KTtcblxuICAgIGlmIChjb25maXJtZWQpIHtcbiAgICAgIGF3YWl0IExvY2FsU3RvcmFnZS5yZW1vdmVJdGVtKFwiY29tbWFuZC1oaXN0b3J5XCIpO1xuICAgICAgc2V0SGlzdG9yeShbXSk7XG4gICAgICBzaG93VG9hc3Qoe1xuICAgICAgICBzdHlsZTogVG9hc3QuU3R5bGUuU3VjY2VzcyxcbiAgICAgICAgdGl0bGU6IFwiXHU1QzY1XHU2Qjc0XHUzMDkyXHUzMEFGXHUzMEVBXHUzMEEyXHUzMDU3XHUzMDdFXHUzMDU3XHUzMDVGXCIsXG4gICAgICB9KTtcbiAgICB9XG4gIH07XG5cbiAgcmV0dXJuIChcbiAgICA8TGlzdCBpc0xvYWRpbmc9e2lzTG9hZGluZ30gc2VhcmNoQmFyUGxhY2Vob2xkZXI9XCJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTY5MUNcdTdEMjIuLi5cIj5cbiAgICAgIDxMaXN0LlNlY3Rpb24gdGl0bGU9XCJcdTMwODhcdTMwNEZcdTRGN0ZcdTMwNDZcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcIj5cbiAgICAgICAge2NvbW1hbmRzLm1hcCgoY29tbWFuZCkgPT4gKFxuICAgICAgICAgIDxMaXN0Lkl0ZW1cbiAgICAgICAgICAgIGtleT17Y29tbWFuZC5pZH1cbiAgICAgICAgICAgIHRpdGxlPXtjb21tYW5kLm5hbWV9XG4gICAgICAgICAgICBzdWJ0aXRsZT17Y29tbWFuZC5jb21tYW5kIHx8IFwiXHUzMEJGXHUzMEZDXHUzMERGXHUzMENBXHUzMEVCXHUzMDkyXHU5NThCXHUzMDRGXCJ9XG4gICAgICAgICAgICBhY2Nlc3Nvcmllcz17W3sgdGV4dDogY29tbWFuZC5kaXJlY3RvcnkgPyBgXHVEODNEXHVEQ0MxICR7Y29tbWFuZC5kaXJlY3Rvcnl9YCA6IFwiXCIgfV19XG4gICAgICAgICAgICBhY3Rpb25zPXtcbiAgICAgICAgICAgICAgPEFjdGlvblBhbmVsPlxuICAgICAgICAgICAgICAgIDxBY3Rpb24gdGl0bGU9XCJcdTVCOUZcdTg4NENcIiBpY29uPXtJY29uLlRlcm1pbmFsfSBvbkFjdGlvbj17KCkgPT4gZXhlY3V0ZUNvbW1hbmQoY29tbWFuZCl9IC8+XG4gICAgICAgICAgICAgICAgPEFjdGlvblxuICAgICAgICAgICAgICAgICAgdGl0bGU9XCJcdTdERThcdTk2QzZcIlxuICAgICAgICAgICAgICAgICAgaWNvbj17SWNvbi5QZW5jaWx9XG4gICAgICAgICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4gcHVzaCg8RWRpdENvbW1hbmRGb3JtIGNvbW1hbmQ9e2NvbW1hbmR9IG9uU2F2ZT17c2F2ZUNvbW1hbmRzfSBjb21tYW5kcz17Y29tbWFuZHN9IC8+KX1cbiAgICAgICAgICAgICAgICAvPlxuICAgICAgICAgICAgICAgIDxBY3Rpb25cbiAgICAgICAgICAgICAgICAgIHRpdGxlPVwiXHU1MjRBXHU5NjY0XCJcbiAgICAgICAgICAgICAgICAgIGljb249e0ljb24uVHJhc2h9XG4gICAgICAgICAgICAgICAgICBzdHlsZT17QWN0aW9uLlN0eWxlLkRlc3RydWN0aXZlfVxuICAgICAgICAgICAgICAgICAgb25BY3Rpb249eygpID0+IGRlbGV0ZUNvbW1hbmQoY29tbWFuZC5pZCl9XG4gICAgICAgICAgICAgICAgLz5cbiAgICAgICAgICAgICAgICA8QWN0aW9uXG4gICAgICAgICAgICAgICAgICB0aXRsZT1cIlx1NjVCMFx1MzA1N1x1MzA0NFx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA5Mlx1OEZGRFx1NTJBMFwiXG4gICAgICAgICAgICAgICAgICBpY29uPXtJY29uLlBsdXN9XG4gICAgICAgICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4gcHVzaCg8QWRkQ29tbWFuZEZvcm0gb25TYXZlPXtzYXZlQ29tbWFuZHN9IGNvbW1hbmRzPXtjb21tYW5kc30gLz4pfVxuICAgICAgICAgICAgICAgICAgc2hvcnRjdXQ9e3sgbW9kaWZpZXJzOiBbXCJjbWRcIl0sIGtleTogXCJuXCIgfX1cbiAgICAgICAgICAgICAgICAvPlxuICAgICAgICAgICAgICAgIDxBY3Rpb25cbiAgICAgICAgICAgICAgICAgIHRpdGxlPVwiXHU1QzY1XHU2Qjc0XHUzMDkyXHU4ODY4XHU3OTNBXCJcbiAgICAgICAgICAgICAgICAgIGljb249e0ljb24uQ2xvY2t9XG4gICAgICAgICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4gcHVzaCg8SGlzdG9yeVZpZXcgaGlzdG9yeT17aGlzdG9yeX0gb25DbGVhcj17Y2xlYXJIaXN0b3J5fSAvPil9XG4gICAgICAgICAgICAgICAgLz5cbiAgICAgICAgICAgICAgPC9BY3Rpb25QYW5lbD5cbiAgICAgICAgICAgIH1cbiAgICAgICAgICAvPlxuICAgICAgICApKX1cbiAgICAgIDwvTGlzdC5TZWN0aW9uPlxuXG4gICAgICB7aGlzdG9yeS5sZW5ndGggPiAwICYmIChcbiAgICAgICAgPExpc3QuU2VjdGlvbiB0aXRsZT1cIlx1NjcwMFx1OEZEMVx1MzA2RVx1NUM2NVx1NkI3NFwiPlxuICAgICAgICAgIHtoaXN0b3J5XG4gICAgICAgICAgICAuc2xpY2UoLTUpXG4gICAgICAgICAgICAucmV2ZXJzZSgpXG4gICAgICAgICAgICAubWFwKChpdGVtKSA9PiAoXG4gICAgICAgICAgICAgIDxMaXN0Lkl0ZW1cbiAgICAgICAgICAgICAgICBrZXk9e2l0ZW0uaWR9XG4gICAgICAgICAgICAgICAgdGl0bGU9e2l0ZW0uY29tbWFuZH1cbiAgICAgICAgICAgICAgICBzdWJ0aXRsZT17YCR7aXRlbS5kaXJlY3Rvcnl9IFx1MjAyMiAke25ldyBEYXRlKGl0ZW0udGltZXN0YW1wKS50b0xvY2FsZVN0cmluZygpfWB9XG4gICAgICAgICAgICAgICAgaWNvbj17SWNvbi5DbG9ja31cbiAgICAgICAgICAgICAgICBhY3Rpb25zPXtcbiAgICAgICAgICAgICAgICAgIDxBY3Rpb25QYW5lbD5cbiAgICAgICAgICAgICAgICAgICAgPEFjdGlvblxuICAgICAgICAgICAgICAgICAgICAgIHRpdGxlPVwiXHU1MThEXHU1QjlGXHU4ODRDXCJcbiAgICAgICAgICAgICAgICAgICAgICBpY29uPXtJY29uLlRlcm1pbmFsfVxuICAgICAgICAgICAgICAgICAgICAgIG9uQWN0aW9uPXsoKSA9PiB7XG4gICAgICAgICAgICAgICAgICAgICAgICBjb25zdCBjbWQ6IENvbW1hbmQgPSB7XG4gICAgICAgICAgICAgICAgICAgICAgICAgIGlkOiBpdGVtLmlkLFxuICAgICAgICAgICAgICAgICAgICAgICAgICBuYW1lOiBpdGVtLmNvbW1hbmQsXG4gICAgICAgICAgICAgICAgICAgICAgICAgIGNvbW1hbmQ6IGl0ZW0uY29tbWFuZCxcbiAgICAgICAgICAgICAgICAgICAgICAgICAgZGlyZWN0b3J5OiBpdGVtLmRpcmVjdG9yeSxcbiAgICAgICAgICAgICAgICAgICAgICAgIH07XG4gICAgICAgICAgICAgICAgICAgICAgICBwdXNoKFxuICAgICAgICAgICAgICAgICAgICAgICAgICA8Q29tbWFuZEV4ZWN1dG9yIGNvbW1hbmQ9e2NtZH0gZGlyZWN0b3J5PXtpdGVtLmRpcmVjdG9yeX0gb25TYXZlSGlzdG9yeT17c2F2ZVRvSGlzdG9yeX0gLz4sXG4gICAgICAgICAgICAgICAgICAgICAgICApO1xuICAgICAgICAgICAgICAgICAgICAgIH19XG4gICAgICAgICAgICAgICAgICAgIC8+XG4gICAgICAgICAgICAgICAgICAgIDxBY3Rpb25cbiAgICAgICAgICAgICAgICAgICAgICB0aXRsZT1cIlx1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQVx1MzA2N1x1OTU4Qlx1MzA0RlwiXG4gICAgICAgICAgICAgICAgICAgICAgaWNvbj17SWNvbi5Gb2xkZXJ9XG4gICAgICAgICAgICAgICAgICAgICAgb25BY3Rpb249eygpID0+IHB1c2goPEludGVyYWN0aXZlVGVybWluYWwgZGlyZWN0b3J5PXtpdGVtLmRpcmVjdG9yeX0gLz4pfVxuICAgICAgICAgICAgICAgICAgICAvPlxuICAgICAgICAgICAgICAgICAgPC9BY3Rpb25QYW5lbD5cbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgIC8+XG4gICAgICAgICAgICApKX1cbiAgICAgICAgPC9MaXN0LlNlY3Rpb24+XG4gICAgICApfVxuICAgIDwvTGlzdD5cbiAgKTtcbn1cblxuZnVuY3Rpb24gQWRkQ29tbWFuZEZvcm0oeyBvblNhdmUsIGNvbW1hbmRzIH06IHsgb25TYXZlOiAoY29tbWFuZHM6IENvbW1hbmRbXSkgPT4gdm9pZDsgY29tbWFuZHM6IENvbW1hbmRbXSB9KSB7XG4gIGNvbnN0IHsgcG9wIH0gPSB1c2VOYXZpZ2F0aW9uKCk7XG4gIGNvbnN0IFtuYW1lRXJyb3IsIHNldE5hbWVFcnJvcl0gPSB1c2VTdGF0ZTxzdHJpbmcgfCB1bmRlZmluZWQ+KCk7XG4gIGNvbnN0IFtjb21tYW5kRXJyb3IsIHNldENvbW1hbmRFcnJvcl0gPSB1c2VTdGF0ZTxzdHJpbmcgfCB1bmRlZmluZWQ+KCk7XG5cbiAgZnVuY3Rpb24gaGFuZGxlU3VibWl0KHZhbHVlczogeyBuYW1lOiBzdHJpbmc7IGNvbW1hbmQ6IHN0cmluZzsgZGVzY3JpcHRpb246IHN0cmluZzsgZGlyZWN0b3J5OiBzdHJpbmcgfSkge1xuICAgIGlmICghdmFsdWVzLm5hbWUudHJpbSgpKSB7XG4gICAgICBzZXROYW1lRXJyb3IoXCJcdTU0MERcdTUyNERcdTMwNkZcdTVGQzVcdTk4MDhcdTMwNjdcdTMwNTlcIik7XG4gICAgICByZXR1cm47XG4gICAgfVxuXG4gICAgY29uc3QgbmV3Q29tbWFuZDogQ29tbWFuZCA9IHtcbiAgICAgIGlkOiBEYXRlLm5vdygpLnRvU3RyaW5nKCksXG4gICAgICBuYW1lOiB2YWx1ZXMubmFtZS50cmltKCksXG4gICAgICBjb21tYW5kOiB2YWx1ZXMuY29tbWFuZC50cmltKCksXG4gICAgICBkZXNjcmlwdGlvbjogdmFsdWVzLmRlc2NyaXB0aW9uLnRyaW0oKSxcbiAgICAgIGRpcmVjdG9yeTogdmFsdWVzLmRpcmVjdG9yeS50cmltKCkgfHwgdW5kZWZpbmVkLFxuICAgIH07XG5cbiAgICBjb25zdCB1cGRhdGVkQ29tbWFuZHMgPSBbLi4uY29tbWFuZHMsIG5ld0NvbW1hbmRdO1xuICAgIG9uU2F2ZSh1cGRhdGVkQ29tbWFuZHMpO1xuXG4gICAgc2hvd1RvYXN0KHtcbiAgICAgIHN0eWxlOiBUb2FzdC5TdHlsZS5TdWNjZXNzLFxuICAgICAgdGl0bGU6IFwiXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XHUzMDkyXHU4RkZEXHU1MkEwXHUzMDU3XHUzMDdFXHUzMDU3XHUzMDVGXCIsXG4gICAgfSk7XG5cbiAgICBwb3AoKTtcbiAgfVxuXG4gIHJldHVybiAoXG4gICAgPEZvcm1cbiAgICAgIGFjdGlvbnM9e1xuICAgICAgICA8QWN0aW9uUGFuZWw+XG4gICAgICAgICAgPEFjdGlvbi5TdWJtaXRGb3JtIHRpdGxlPVwiXHU4RkZEXHU1MkEwXCIgb25TdWJtaXQ9e2hhbmRsZVN1Ym1pdH0gLz5cbiAgICAgICAgICA8QWN0aW9uIHRpdGxlPVwiXHUzMEFEXHUzMEUzXHUzMEYzXHUzMEJCXHUzMEVCXCIgb25BY3Rpb249e3BvcH0gLz5cbiAgICAgICAgPC9BY3Rpb25QYW5lbD5cbiAgICAgIH1cbiAgICA+XG4gICAgICA8Rm9ybS5UZXh0RmllbGRcbiAgICAgICAgaWQ9XCJuYW1lXCJcbiAgICAgICAgdGl0bGU9XCJcdTU0MERcdTUyNERcIlxuICAgICAgICBwbGFjZWhvbGRlcj1cIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA2RVx1NTQwRFx1NTI0RFx1MzA5Mlx1NTE2NVx1NTI5QlwiXG4gICAgICAgIGVycm9yPXtuYW1lRXJyb3J9XG4gICAgICAgIG9uQ2hhbmdlPXsoKSA9PiBzZXROYW1lRXJyb3IodW5kZWZpbmVkKX1cbiAgICAgIC8+XG4gICAgICA8Rm9ybS5UZXh0RmllbGRcbiAgICAgICAgaWQ9XCJjb21tYW5kXCJcbiAgICAgICAgdGl0bGU9XCJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcIlxuICAgICAgICBwbGFjZWhvbGRlcj1cIlx1NUI5Rlx1ODg0Q1x1MzA1OVx1MzA4Qlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA5Mlx1NTE2NVx1NTI5Qlx1RkYwOFx1N0E3QVx1MzA2RVx1NTgzNFx1NTQwOFx1MzA2Rlx1MzBCRlx1MzBGQ1x1MzBERlx1MzBDQVx1MzBFQlx1MzA2RVx1MzA3Rlx1OTU4Qlx1MzA0Rlx1RkYwOVwiXG4gICAgICAgIGVycm9yPXtjb21tYW5kRXJyb3J9XG4gICAgICAgIG9uQ2hhbmdlPXsoKSA9PiBzZXRDb21tYW5kRXJyb3IodW5kZWZpbmVkKX1cbiAgICAgIC8+XG4gICAgICA8Rm9ybS5UZXh0RmllbGQgaWQ9XCJkZXNjcmlwdGlvblwiIHRpdGxlPVwiXHU4QUFDXHU2NjBFXCIgcGxhY2Vob2xkZXI9XCJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwNkVcdThBQUNcdTY2MEVcdUZGMDhcdTMwQUFcdTMwRDdcdTMwQjdcdTMwRTdcdTMwRjNcdUZGMDlcIiAvPlxuICAgICAgPEZvcm0uVGV4dEZpZWxkIGlkPVwiZGlyZWN0b3J5XCIgdGl0bGU9XCJcdTMwQzdcdTMwQTNcdTMwRUNcdTMwQUZcdTMwQzhcdTMwRUFcIiBwbGFjZWhvbGRlcj1cIlx1NUI5Rlx1ODg0Q1x1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQVx1RkYwOFx1N0E3QVx1MzA2RVx1NTgzNFx1NTQwOFx1MzA2Rlx1MzBDN1x1MzBENVx1MzBBOVx1MzBFQlx1MzBDOFx1RkYwOVwiIC8+XG4gICAgPC9Gb3JtPlxuICApO1xufVxuXG5mdW5jdGlvbiBFZGl0Q29tbWFuZEZvcm0oe1xuICBjb21tYW5kLFxuICBvblNhdmUsXG4gIGNvbW1hbmRzLFxufToge1xuICBjb21tYW5kOiBDb21tYW5kO1xuICBvblNhdmU6IChjb21tYW5kczogQ29tbWFuZFtdKSA9PiB2b2lkO1xuICBjb21tYW5kczogQ29tbWFuZFtdO1xufSkge1xuICBjb25zdCB7IHBvcCB9ID0gdXNlTmF2aWdhdGlvbigpO1xuICBjb25zdCBbbmFtZUVycm9yLCBzZXROYW1lRXJyb3JdID0gdXNlU3RhdGU8c3RyaW5nIHwgdW5kZWZpbmVkPigpO1xuXG4gIGZ1bmN0aW9uIGhhbmRsZVN1Ym1pdCh2YWx1ZXM6IHsgbmFtZTogc3RyaW5nOyBjb21tYW5kOiBzdHJpbmc7IGRlc2NyaXB0aW9uOiBzdHJpbmc7IGRpcmVjdG9yeTogc3RyaW5nIH0pIHtcbiAgICBpZiAoIXZhbHVlcy5uYW1lLnRyaW0oKSkge1xuICAgICAgc2V0TmFtZUVycm9yKFwiXHU1NDBEXHU1MjREXHUzMDZGXHU1RkM1XHU5ODA4XHUzMDY3XHUzMDU5XCIpO1xuICAgICAgcmV0dXJuO1xuICAgIH1cblxuICAgIGNvbnN0IHVwZGF0ZWRDb21tYW5kOiBDb21tYW5kID0ge1xuICAgICAgLi4uY29tbWFuZCxcbiAgICAgIG5hbWU6IHZhbHVlcy5uYW1lLnRyaW0oKSxcbiAgICAgIGNvbW1hbmQ6IHZhbHVlcy5jb21tYW5kLnRyaW0oKSxcbiAgICAgIGRlc2NyaXB0aW9uOiB2YWx1ZXMuZGVzY3JpcHRpb24udHJpbSgpLFxuICAgICAgZGlyZWN0b3J5OiB2YWx1ZXMuZGlyZWN0b3J5LnRyaW0oKSB8fCB1bmRlZmluZWQsXG4gICAgfTtcblxuICAgIGNvbnN0IHVwZGF0ZWRDb21tYW5kcyA9IGNvbW1hbmRzLm1hcCgoY21kKSA9PiAoY21kLmlkID09PSBjb21tYW5kLmlkID8gdXBkYXRlZENvbW1hbmQgOiBjbWQpKTtcblxuICAgIG9uU2F2ZSh1cGRhdGVkQ29tbWFuZHMpO1xuXG4gICAgc2hvd1RvYXN0KHtcbiAgICAgIHN0eWxlOiBUb2FzdC5TdHlsZS5TdWNjZXNzLFxuICAgICAgdGl0bGU6IFwiXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XHUzMDkyXHU2NkY0XHU2NUIwXHUzMDU3XHUzMDdFXHUzMDU3XHUzMDVGXCIsXG4gICAgfSk7XG5cbiAgICBwb3AoKTtcbiAgfVxuXG4gIHJldHVybiAoXG4gICAgPEZvcm1cbiAgICAgIGFjdGlvbnM9e1xuICAgICAgICA8QWN0aW9uUGFuZWw+XG4gICAgICAgICAgPEFjdGlvbi5TdWJtaXRGb3JtIHRpdGxlPVwiXHU0RkREXHU1QjU4XCIgb25TdWJtaXQ9e2hhbmRsZVN1Ym1pdH0gLz5cbiAgICAgICAgICA8QWN0aW9uIHRpdGxlPVwiXHUzMEFEXHUzMEUzXHUzMEYzXHUzMEJCXHUzMEVCXCIgb25BY3Rpb249e3BvcH0gLz5cbiAgICAgICAgPC9BY3Rpb25QYW5lbD5cbiAgICAgIH1cbiAgICA+XG4gICAgICA8Rm9ybS5UZXh0RmllbGRcbiAgICAgICAgaWQ9XCJuYW1lXCJcbiAgICAgICAgdGl0bGU9XCJcdTU0MERcdTUyNERcIlxuICAgICAgICBkZWZhdWx0VmFsdWU9e2NvbW1hbmQubmFtZX1cbiAgICAgICAgZXJyb3I9e25hbWVFcnJvcn1cbiAgICAgICAgb25DaGFuZ2U9eygpID0+IHNldE5hbWVFcnJvcih1bmRlZmluZWQpfVxuICAgICAgLz5cbiAgICAgIDxGb3JtLlRleHRGaWVsZCBpZD1cImNvbW1hbmRcIiB0aXRsZT1cIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVwiIGRlZmF1bHRWYWx1ZT17Y29tbWFuZC5jb21tYW5kfSAvPlxuICAgICAgPEZvcm0uVGV4dEZpZWxkIGlkPVwiZGVzY3JpcHRpb25cIiB0aXRsZT1cIlx1OEFBQ1x1NjYwRVwiIGRlZmF1bHRWYWx1ZT17Y29tbWFuZC5kZXNjcmlwdGlvbiB8fCBcIlwifSAvPlxuICAgICAgPEZvcm0uVGV4dEZpZWxkIGlkPVwiZGlyZWN0b3J5XCIgdGl0bGU9XCJcdTMwQzdcdTMwQTNcdTMwRUNcdTMwQUZcdTMwQzhcdTMwRUFcIiBkZWZhdWx0VmFsdWU9e2NvbW1hbmQuZGlyZWN0b3J5IHx8IFwiXCJ9IC8+XG4gICAgPC9Gb3JtPlxuICApO1xufVxuXG5mdW5jdGlvbiBIaXN0b3J5Vmlldyh7IGhpc3RvcnksIG9uQ2xlYXIgfTogeyBoaXN0b3J5OiBDb21tYW5kSGlzdG9yeVtdOyBvbkNsZWFyOiAoKSA9PiB2b2lkIH0pIHtcbiAgY29uc3QgeyBwdXNoIH0gPSB1c2VOYXZpZ2F0aW9uKCk7XG5cbiAgcmV0dXJuIChcbiAgICA8TGlzdCBzZWFyY2hCYXJQbGFjZWhvbGRlcj1cIlx1NUM2NVx1NkI3NFx1MzA5Mlx1NjkxQ1x1N0QyMi4uLlwiPlxuICAgICAgPExpc3QuU2VjdGlvbiB0aXRsZT17YFx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1NUM2NVx1NkI3NCAoJHtoaXN0b3J5Lmxlbmd0aH1cdTRFRjYpYH0+XG4gICAgICAgIHtoaXN0b3J5XG4gICAgICAgICAgLnNsaWNlKClcbiAgICAgICAgICAucmV2ZXJzZSgpXG4gICAgICAgICAgLm1hcCgoaXRlbSkgPT4gKFxuICAgICAgICAgICAgPExpc3QuSXRlbVxuICAgICAgICAgICAgICBrZXk9e2l0ZW0uaWR9XG4gICAgICAgICAgICAgIHRpdGxlPXtpdGVtLmNvbW1hbmR9XG4gICAgICAgICAgICAgIHN1YnRpdGxlPXtpdGVtLmRpcmVjdG9yeX1cbiAgICAgICAgICAgICAgYWNjZXNzb3JpZXM9e1t7IHRleHQ6IG5ldyBEYXRlKGl0ZW0udGltZXN0YW1wKS50b0xvY2FsZVN0cmluZygpIH1dfVxuICAgICAgICAgICAgICBhY3Rpb25zPXtcbiAgICAgICAgICAgICAgICA8QWN0aW9uUGFuZWw+XG4gICAgICAgICAgICAgICAgICA8QWN0aW9uXG4gICAgICAgICAgICAgICAgICAgIHRpdGxlPVwiXHU1MThEXHU1QjlGXHU4ODRDXCJcbiAgICAgICAgICAgICAgICAgICAgaWNvbj17SWNvbi5UZXJtaW5hbH1cbiAgICAgICAgICAgICAgICAgICAgb25BY3Rpb249eygpID0+IHtcbiAgICAgICAgICAgICAgICAgICAgICBjb25zdCBjbWQ6IENvbW1hbmQgPSB7XG4gICAgICAgICAgICAgICAgICAgICAgICBpZDogaXRlbS5pZCxcbiAgICAgICAgICAgICAgICAgICAgICAgIG5hbWU6IGl0ZW0uY29tbWFuZCxcbiAgICAgICAgICAgICAgICAgICAgICAgIGNvbW1hbmQ6IGl0ZW0uY29tbWFuZCxcbiAgICAgICAgICAgICAgICAgICAgICAgIGRpcmVjdG9yeTogaXRlbS5kaXJlY3RvcnksXG4gICAgICAgICAgICAgICAgICAgICAgfTtcbiAgICAgICAgICAgICAgICAgICAgICBwdXNoKDxDb21tYW5kRXhlY3V0b3IgY29tbWFuZD17Y21kfSBkaXJlY3Rvcnk9e2l0ZW0uZGlyZWN0b3J5fSAvPik7XG4gICAgICAgICAgICAgICAgICAgIH19XG4gICAgICAgICAgICAgICAgICAvPlxuICAgICAgICAgICAgICAgICAgPEFjdGlvblxuICAgICAgICAgICAgICAgICAgICB0aXRsZT1cIlx1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQVx1MzA2N1x1OTU4Qlx1MzA0RlwiXG4gICAgICAgICAgICAgICAgICAgIGljb249e0ljb24uRm9sZGVyfVxuICAgICAgICAgICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4gcHVzaCg8SW50ZXJhY3RpdmVUZXJtaW5hbCBkaXJlY3Rvcnk9e2l0ZW0uZGlyZWN0b3J5fSAvPil9XG4gICAgICAgICAgICAgICAgICAvPlxuICAgICAgICAgICAgICAgICAgPEFjdGlvbiB0aXRsZT1cIlx1NUM2NVx1NkI3NFx1MzA5Mlx1MzBBRlx1MzBFQVx1MzBBMlwiIGljb249e0ljb24uVHJhc2h9IHN0eWxlPXtBY3Rpb24uU3R5bGUuRGVzdHJ1Y3RpdmV9IG9uQWN0aW9uPXtvbkNsZWFyfSAvPlxuICAgICAgICAgICAgICAgIDwvQWN0aW9uUGFuZWw+XG4gICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIC8+XG4gICAgICAgICAgKSl9XG4gICAgICA8L0xpc3QuU2VjdGlvbj5cbiAgICA8L0xpc3Q+XG4gICk7XG59XG5cbi8vIFx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1NUI5Rlx1ODg0Q1x1MzBCM1x1MzBGM1x1MzBERFx1MzBGQ1x1MzBDRFx1MzBGM1x1MzBDOFxuZnVuY3Rpb24gQ29tbWFuZEV4ZWN1dG9yKHtcbiAgY29tbWFuZCxcbiAgZGlyZWN0b3J5LFxuICBvblNhdmVIaXN0b3J5LFxufToge1xuICBjb21tYW5kOiBDb21tYW5kO1xuICBkaXJlY3Rvcnk6IHN0cmluZztcbiAgb25TYXZlSGlzdG9yeT86IChjb21tYW5kOiBzdHJpbmcsIGRpcmVjdG9yeTogc3RyaW5nKSA9PiB2b2lkO1xufSkge1xuICBjb25zdCBbcmVzdWx0LCBzZXRSZXN1bHRdID0gdXNlU3RhdGU8Q29tbWFuZFJlc3VsdCB8IG51bGw+KG51bGwpO1xuICBjb25zdCBbaXNMb2FkaW5nLCBzZXRJc0xvYWRpbmddID0gdXNlU3RhdGUodHJ1ZSk7XG4gIGNvbnN0IHsgcG9wIH0gPSB1c2VOYXZpZ2F0aW9uKCk7XG5cbiAgdXNlRWZmZWN0KCgpID0+IHtcbiAgICBleGVjdXRlQ29tbWFuZCgpO1xuICB9LCBbXSk7XG5cbiAgY29uc3QgZXhlY3V0ZUNvbW1hbmQgPSBhc3luYyAoKSA9PiB7XG4gICAgY29uc3Qgc3RhcnRUaW1lID0gRGF0ZS5ub3coKTtcblxuICAgIHRyeSB7XG4gICAgICBjb25zdCB7IHN0ZG91dCwgc3RkZXJyIH0gPSBhd2FpdCBleGVjQXN5bmMoY29tbWFuZC5jb21tYW5kLCB7XG4gICAgICAgIGN3ZDogZGlyZWN0b3J5LFxuICAgICAgICBlbnY6IHsgLi4ucHJvY2Vzcy5lbnYsIFBXRDogZGlyZWN0b3J5IH0sXG4gICAgICAgIG1heEJ1ZmZlcjogMTAyNCAqIDEwMjQgKiAxMCwgLy8gMTBNQlxuICAgICAgfSk7XG5cbiAgICAgIGNvbnN0IGVuZFRpbWUgPSBEYXRlLm5vdygpO1xuICAgICAgY29uc3QgY29tbWFuZFJlc3VsdDogQ29tbWFuZFJlc3VsdCA9IHtcbiAgICAgICAgaWQ6IERhdGUubm93KCkudG9TdHJpbmcoKSxcbiAgICAgICAgY29tbWFuZDogY29tbWFuZC5jb21tYW5kLFxuICAgICAgICBkaXJlY3RvcnksXG4gICAgICAgIG91dHB1dDogc3Rkb3V0LFxuICAgICAgICBlcnJvcjogc3RkZXJyLFxuICAgICAgICBleGl0Q29kZTogMCxcbiAgICAgICAgdGltZXN0YW1wOiBzdGFydFRpbWUsXG4gICAgICAgIGR1cmF0aW9uOiBlbmRUaW1lIC0gc3RhcnRUaW1lLFxuICAgICAgfTtcblxuICAgICAgc2V0UmVzdWx0KGNvbW1hbmRSZXN1bHQpO1xuXG4gICAgICBpZiAob25TYXZlSGlzdG9yeSkge1xuICAgICAgICBhd2FpdCBvblNhdmVIaXN0b3J5KGNvbW1hbmQuY29tbWFuZCwgZGlyZWN0b3J5KTtcbiAgICAgIH1cbiAgICB9IGNhdGNoIChlcnJvcjogdW5rbm93bikge1xuICAgICAgY29uc3QgZW5kVGltZSA9IERhdGUubm93KCk7XG4gICAgICBjb25zdCBlcnJvck9iaiA9IGVycm9yIGFzIHsgc3Rkb3V0Pzogc3RyaW5nOyBzdGRlcnI/OiBzdHJpbmc7IG1lc3NhZ2U/OiBzdHJpbmc7IGNvZGU/OiBudW1iZXIgfTtcbiAgICAgIGNvbnN0IGNvbW1hbmRSZXN1bHQ6IENvbW1hbmRSZXN1bHQgPSB7XG4gICAgICAgIGlkOiBEYXRlLm5vdygpLnRvU3RyaW5nKCksXG4gICAgICAgIGNvbW1hbmQ6IGNvbW1hbmQuY29tbWFuZCxcbiAgICAgICAgZGlyZWN0b3J5LFxuICAgICAgICBvdXRwdXQ6IGVycm9yT2JqLnN0ZG91dCB8fCBcIlwiLFxuICAgICAgICBlcnJvcjogZXJyb3JPYmouc3RkZXJyIHx8IGVycm9yT2JqLm1lc3NhZ2UgfHwgXCJVbmtub3duIGVycm9yXCIsXG4gICAgICAgIGV4aXRDb2RlOiBlcnJvck9iai5jb2RlIHx8IDEsXG4gICAgICAgIHRpbWVzdGFtcDogc3RhcnRUaW1lLFxuICAgICAgICBkdXJhdGlvbjogZW5kVGltZSAtIHN0YXJ0VGltZSxcbiAgICAgIH07XG5cbiAgICAgIHNldFJlc3VsdChjb21tYW5kUmVzdWx0KTtcblxuICAgICAgaWYgKG9uU2F2ZUhpc3RvcnkpIHtcbiAgICAgICAgYXdhaXQgb25TYXZlSGlzdG9yeShjb21tYW5kLmNvbW1hbmQsIGRpcmVjdG9yeSk7XG4gICAgICB9XG4gICAgfSBmaW5hbGx5IHtcbiAgICAgIHNldElzTG9hZGluZyhmYWxzZSk7XG4gICAgfVxuICB9O1xuXG4gIGNvbnN0IGdldE1hcmtkb3duID0gKCkgPT4ge1xuICAgIGlmICghcmVzdWx0KSByZXR1cm4gXCJcIjtcblxuICAgIGxldCBtYXJrZG93biA9IGAjIFx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1NUI5Rlx1ODg0Q1x1N0Q1MFx1Njc5Q1xuXG4qKlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOToqKiBcXGAke3Jlc3VsdC5jb21tYW5kfVxcYCAgXG4qKlx1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQToqKiBcXGAke3Jlc3VsdC5kaXJlY3Rvcnl9XFxgICBcbioqXHU1QjlGXHU4ODRDXHU2NjQyXHU5NTkzOioqICR7cmVzdWx0LmR1cmF0aW9ufW1zICBcbioqXHU3RDQyXHU0RTg2XHUzMEIzXHUzMEZDXHUzMEM5OioqICR7cmVzdWx0LmV4aXRDb2RlfSAgXG4qKlx1NUI5Rlx1ODg0Q1x1NjVFNVx1NjY0MjoqKiAke25ldyBEYXRlKHJlc3VsdC50aW1lc3RhbXApLnRvTG9jYWxlU3RyaW5nKCl9XG5cbmA7XG5cbiAgICBpZiAocmVzdWx0Lm91dHB1dCkge1xuICAgICAgbWFya2Rvd24gKz0gYCMjIFx1NTFGQVx1NTI5QlxuXG5cXGBcXGBcXGBcbiR7cmVzdWx0Lm91dHB1dH1cblxcYFxcYFxcYFxuXG5gO1xuICAgIH1cblxuICAgIGlmIChyZXN1bHQuZXJyb3IpIHtcbiAgICAgIG1hcmtkb3duICs9IGAjIyBcdTMwQThcdTMwRTlcdTMwRkNcblxuXFxgXFxgXFxgXG4ke3Jlc3VsdC5lcnJvcn1cblxcYFxcYFxcYFxuXG5gO1xuICAgIH1cblxuICAgIHJldHVybiBtYXJrZG93bjtcbiAgfTtcblxuICByZXR1cm4gKFxuICAgIDxEZXRhaWxcbiAgICAgIGlzTG9hZGluZz17aXNMb2FkaW5nfVxuICAgICAgbWFya2Rvd249e2dldE1hcmtkb3duKCl9XG4gICAgICBhY3Rpb25zPXtcbiAgICAgICAgPEFjdGlvblBhbmVsPlxuICAgICAgICAgIDxBY3Rpb25cbiAgICAgICAgICAgIHRpdGxlPVwiXHU1MThEXHU1QjlGXHU4ODRDXCJcbiAgICAgICAgICAgIGljb249e0ljb24uQXJyb3dDbG9ja3dpc2V9XG4gICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4ge1xuICAgICAgICAgICAgICBzZXRJc0xvYWRpbmcodHJ1ZSk7XG4gICAgICAgICAgICAgIGV4ZWN1dGVDb21tYW5kKCk7XG4gICAgICAgICAgICB9fVxuICAgICAgICAgIC8+XG4gICAgICAgICAgPEFjdGlvbiB0aXRsZT1cIlx1NjIzQlx1MzA4QlwiIGljb249e0ljb24uQXJyb3dMZWZ0fSBvbkFjdGlvbj17cG9wfSBzaG9ydGN1dD17eyBtb2RpZmllcnM6IFtcImNtZFwiXSwga2V5OiBcIndcIiB9fSAvPlxuICAgICAgICAgIDxBY3Rpb24uQ29weVRvQ2xpcGJvYXJkXG4gICAgICAgICAgICB0aXRsZT1cIlx1NTFGQVx1NTI5Qlx1MzA5Mlx1MzBCM1x1MzBENFx1MzBGQ1wiXG4gICAgICAgICAgICBjb250ZW50PXtyZXN1bHQ/Lm91dHB1dCB8fCBcIlwifVxuICAgICAgICAgICAgc2hvcnRjdXQ9e3sgbW9kaWZpZXJzOiBbXCJjbWRcIl0sIGtleTogXCJjXCIgfX1cbiAgICAgICAgICAvPlxuICAgICAgICAgIDxBY3Rpb24uQ29weVRvQ2xpcGJvYXJkXG4gICAgICAgICAgICB0aXRsZT1cIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA5Mlx1MzBCM1x1MzBENFx1MzBGQ1wiXG4gICAgICAgICAgICBjb250ZW50PXtyZXN1bHQ/LmNvbW1hbmQgfHwgXCJcIn1cbiAgICAgICAgICAgIHNob3J0Y3V0PXt7IG1vZGlmaWVyczogW1wiY21kXCIsIFwic2hpZnRcIl0sIGtleTogXCJjXCIgfX1cbiAgICAgICAgICAvPlxuICAgICAgICA8L0FjdGlvblBhbmVsPlxuICAgICAgfVxuICAgIC8+XG4gICk7XG59XG5cbi8vIFx1MzBBNFx1MzBGM1x1MzBCRlx1MzBFOVx1MzBBRlx1MzBDNlx1MzBBM1x1MzBENlx1MzBCRlx1MzBGQ1x1MzBERlx1MzBDQVx1MzBFQlx1MzBCM1x1MzBGM1x1MzBERFx1MzBGQ1x1MzBDRFx1MzBGM1x1MzBDOFxuZnVuY3Rpb24gSW50ZXJhY3RpdmVUZXJtaW5hbCh7IGRpcmVjdG9yeSB9OiB7IGRpcmVjdG9yeTogc3RyaW5nIH0pIHtcbiAgY29uc3QgW2NvbW1hbmRIaXN0b3J5LCBzZXRDb21tYW5kSGlzdG9yeV0gPSB1c2VTdGF0ZTxzdHJpbmdbXT4oW10pO1xuICBjb25zdCBbcmVzdWx0cywgc2V0UmVzdWx0c10gPSB1c2VTdGF0ZTxDb21tYW5kUmVzdWx0W10+KFtdKTtcbiAgY29uc3QgeyBwb3AsIHB1c2ggfSA9IHVzZU5hdmlnYXRpb24oKTtcblxuICBjb25zdCBleGVjdXRlQ29tbWFuZCA9IGFzeW5jIChjb21tYW5kOiBzdHJpbmcpID0+IHtcbiAgICBpZiAoIWNvbW1hbmQudHJpbSgpKSByZXR1cm47XG5cbiAgICBjb25zdCBzdGFydFRpbWUgPSBEYXRlLm5vdygpO1xuXG4gICAgdHJ5IHtcbiAgICAgIGNvbnN0IHsgc3Rkb3V0LCBzdGRlcnIgfSA9IGF3YWl0IGV4ZWNBc3luYyhjb21tYW5kLCB7XG4gICAgICAgIGN3ZDogZGlyZWN0b3J5LFxuICAgICAgICBlbnY6IHsgLi4ucHJvY2Vzcy5lbnYsIFBXRDogZGlyZWN0b3J5IH0sXG4gICAgICAgIG1heEJ1ZmZlcjogMTAyNCAqIDEwMjQgKiAxMCxcbiAgICAgIH0pO1xuXG4gICAgICBjb25zdCBlbmRUaW1lID0gRGF0ZS5ub3coKTtcbiAgICAgIGNvbnN0IHJlc3VsdDogQ29tbWFuZFJlc3VsdCA9IHtcbiAgICAgICAgaWQ6IERhdGUubm93KCkudG9TdHJpbmcoKSxcbiAgICAgICAgY29tbWFuZCxcbiAgICAgICAgZGlyZWN0b3J5LFxuICAgICAgICBvdXRwdXQ6IHN0ZG91dCxcbiAgICAgICAgZXJyb3I6IHN0ZGVycixcbiAgICAgICAgZXhpdENvZGU6IDAsXG4gICAgICAgIHRpbWVzdGFtcDogc3RhcnRUaW1lLFxuICAgICAgICBkdXJhdGlvbjogZW5kVGltZSAtIHN0YXJ0VGltZSxcbiAgICAgIH07XG5cbiAgICAgIHNldFJlc3VsdHMoKHByZXYpID0+IFsuLi5wcmV2LCByZXN1bHRdKTtcbiAgICAgIHNldENvbW1hbmRIaXN0b3J5KChwcmV2KSA9PiBbLi4ucHJldiwgY29tbWFuZF0pO1xuICAgIH0gY2F0Y2ggKGVycm9yOiB1bmtub3duKSB7XG4gICAgICBjb25zdCBlbmRUaW1lID0gRGF0ZS5ub3coKTtcbiAgICAgIGNvbnN0IGVycm9yT2JqID0gZXJyb3IgYXMgeyBzdGRvdXQ/OiBzdHJpbmc7IHN0ZGVycj86IHN0cmluZzsgbWVzc2FnZT86IHN0cmluZzsgY29kZT86IG51bWJlciB9O1xuICAgICAgY29uc3QgcmVzdWx0OiBDb21tYW5kUmVzdWx0ID0ge1xuICAgICAgICBpZDogRGF0ZS5ub3coKS50b1N0cmluZygpLFxuICAgICAgICBjb21tYW5kLFxuICAgICAgICBkaXJlY3RvcnksXG4gICAgICAgIG91dHB1dDogZXJyb3JPYmouc3Rkb3V0IHx8IFwiXCIsXG4gICAgICAgIGVycm9yOiBlcnJvck9iai5zdGRlcnIgfHwgZXJyb3JPYmoubWVzc2FnZSB8fCBcIlVua25vd24gZXJyb3JcIixcbiAgICAgICAgZXhpdENvZGU6IGVycm9yT2JqLmNvZGUgfHwgMSxcbiAgICAgICAgdGltZXN0YW1wOiBzdGFydFRpbWUsXG4gICAgICAgIGR1cmF0aW9uOiBlbmRUaW1lIC0gc3RhcnRUaW1lLFxuICAgICAgfTtcblxuICAgICAgc2V0UmVzdWx0cygocHJldikgPT4gWy4uLnByZXYsIHJlc3VsdF0pO1xuICAgICAgc2V0Q29tbWFuZEhpc3RvcnkoKHByZXYpID0+IFsuLi5wcmV2LCBjb21tYW5kXSk7XG4gICAgfVxuICB9O1xuXG4gIGNvbnN0IGdldFRlcm1pbmFsTWFya2Rvd24gPSAoKSA9PiB7XG4gICAgbGV0IG1hcmtkb3duID0gYCMgXHUzMEE0XHUzMEYzXHUzMEJGXHUzMEU5XHUzMEFGXHUzMEM2XHUzMEEzXHUzMEQ2XHUzMEJGXHUzMEZDXHUzMERGXHUzMENBXHUzMEVCXG5cbioqXHU3M0ZFXHU1NzI4XHUzMDZFXHUzMEM3XHUzMEEzXHUzMEVDXHUzMEFGXHUzMEM4XHUzMEVBOioqIFxcYCR7ZGlyZWN0b3J5fVxcYFxuXG4tLS1cblxuYDtcblxuICAgIHJlc3VsdHMuZm9yRWFjaCgocmVzdWx0KSA9PiB7XG4gICAgICBtYXJrZG93biArPSBgIyMgXFxgJHtyZXN1bHQuY29tbWFuZH1cXGBcbioqXHU1QjlGXHU4ODRDXHU2NjQyXHU5NTkzOioqICR7cmVzdWx0LmR1cmF0aW9ufW1zIHwgKipcdTdENDJcdTRFODZcdTMwQjNcdTMwRkNcdTMwQzk6KiogJHtyZXN1bHQuZXhpdENvZGV9IHwgKipcdTVCOUZcdTg4NENcdTY2NDJcdTUyM0I6KiogJHtuZXcgRGF0ZShyZXN1bHQudGltZXN0YW1wKS50b0xvY2FsZVRpbWVTdHJpbmcoKX1cblxuYDtcblxuICAgICAgaWYgKHJlc3VsdC5vdXRwdXQpIHtcbiAgICAgICAgbWFya2Rvd24gKz0gYFxcYFxcYFxcYFxuJHtyZXN1bHQub3V0cHV0fVxuXFxgXFxgXFxgXG5cbmA7XG4gICAgICB9XG5cbiAgICAgIGlmIChyZXN1bHQuZXJyb3IpIHtcbiAgICAgICAgbWFya2Rvd24gKz0gYCoqXHUzMEE4XHUzMEU5XHUzMEZDOioqXG5cXGBcXGBcXGBcbiR7cmVzdWx0LmVycm9yfVxuXFxgXFxgXFxgXG5cbmA7XG4gICAgICB9XG5cbiAgICAgIG1hcmtkb3duICs9IFwiLS0tXFxuXFxuXCI7XG4gICAgfSk7XG5cbiAgICBpZiAocmVzdWx0cy5sZW5ndGggPT09IDApIHtcbiAgICAgIG1hcmtkb3duICs9IGBcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTUxNjVcdTUyOUJcdTMwNTdcdTMwNjZcdTVCOUZcdTg4NENcdTMwNTdcdTMwNjZcdTMwNEZcdTMwNjBcdTMwNTVcdTMwNDRcdTMwMDJcblxuKipcdTRGN0ZcdTc1MjhcdTRGOEI6Kipcbi0gXFxgbHMgLWxhXFxgIC0gXHUzMEQ1XHUzMEExXHUzMEE0XHUzMEVCXHU0RTAwXHU4OUE3XHU4ODY4XHU3OTNBXG4tIFxcYHB3ZFxcYCAtIFx1NzNGRVx1NTcyOFx1MzA2RVx1MzBDN1x1MzBBM1x1MzBFQ1x1MzBBRlx1MzBDOFx1MzBFQVx1ODg2OFx1NzkzQVxuLSBcXGBnaXQgc3RhdHVzXFxgIC0gR2l0XHU3MkI2XHU2MTRCXHU3OEJBXHU4QThEXG4tIFxcYG5wbSBpbnN0YWxsXFxgIC0gXHUzMEQxXHUzMEMzXHUzMEIxXHUzMEZDXHUzMEI4XHUzMEE0XHUzMEYzXHUzMEI5XHUzMEM4XHUzMEZDXHUzMEVCXG5cbmA7XG4gICAgfVxuXG4gICAgcmV0dXJuIG1hcmtkb3duO1xuICB9O1xuXG4gIHJldHVybiAoXG4gICAgPERldGFpbFxuICAgICAgbWFya2Rvd249e2dldFRlcm1pbmFsTWFya2Rvd24oKX1cbiAgICAgIGFjdGlvbnM9e1xuICAgICAgICA8QWN0aW9uUGFuZWw+XG4gICAgICAgICAgPEFjdGlvblxuICAgICAgICAgICAgdGl0bGU9XCJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTUxNjVcdTUyOUJcIlxuICAgICAgICAgICAgaWNvbj17SWNvbi5UZXJtaW5hbH1cbiAgICAgICAgICAgIG9uQWN0aW9uPXsoKSA9PiB7XG4gICAgICAgICAgICAgIHB1c2goPENvbW1hbmRJbnB1dEZvcm0gb25FeGVjdXRlPXtleGVjdXRlQ29tbWFuZH0gZGlyZWN0b3J5PXtkaXJlY3Rvcnl9IC8+KTtcbiAgICAgICAgICAgIH19XG4gICAgICAgICAgLz5cbiAgICAgICAgICA8QWN0aW9uIHRpdGxlPVwiXHU2MjNCXHUzMDhCXCIgaWNvbj17SWNvbi5BcnJvd0xlZnR9IG9uQWN0aW9uPXtwb3B9IHNob3J0Y3V0PXt7IG1vZGlmaWVyczogW1wiY21kXCJdLCBrZXk6IFwid1wiIH19IC8+XG4gICAgICAgICAgPEFjdGlvblxuICAgICAgICAgICAgdGl0bGU9XCJcdTVDNjVcdTZCNzRcdTMwOTJcdTMwQUZcdTMwRUFcdTMwQTJcIlxuICAgICAgICAgICAgaWNvbj17SWNvbi5UcmFzaH1cbiAgICAgICAgICAgIHN0eWxlPXtBY3Rpb24uU3R5bGUuRGVzdHJ1Y3RpdmV9XG4gICAgICAgICAgICBvbkFjdGlvbj17KCkgPT4ge1xuICAgICAgICAgICAgICBzZXRSZXN1bHRzKFtdKTtcbiAgICAgICAgICAgICAgc2V0Q29tbWFuZEhpc3RvcnkoW10pO1xuICAgICAgICAgICAgfX1cbiAgICAgICAgICAvPlxuICAgICAgICA8L0FjdGlvblBhbmVsPlxuICAgICAgfVxuICAgICAgbWV0YWRhdGE9e1xuICAgICAgICA8RGV0YWlsLk1ldGFkYXRhPlxuICAgICAgICAgIDxEZXRhaWwuTWV0YWRhdGEuVGFnTGlzdCB0aXRsZT1cIlx1NUI5Rlx1ODg0Q1x1NkUwOFx1MzA3Rlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVwiPlxuICAgICAgICAgICAge2NvbW1hbmRIaXN0b3J5LnNsaWNlKC01KS5tYXAoKGNtZCwgaW5kZXgpID0+IChcbiAgICAgICAgICAgICAgPERldGFpbC5NZXRhZGF0YS5UYWdMaXN0Lkl0ZW0ga2V5PXtpbmRleH0gdGV4dD17Y21kfSAvPlxuICAgICAgICAgICAgKSl9XG4gICAgICAgICAgPC9EZXRhaWwuTWV0YWRhdGEuVGFnTGlzdD5cbiAgICAgICAgPC9EZXRhaWwuTWV0YWRhdGE+XG4gICAgICB9XG4gICAgLz5cbiAgKTtcbn1cblxuLy8gXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XHU1MTY1XHU1MjlCXHUzMEQ1XHUzMEE5XHUzMEZDXHUzMEUwXG5mdW5jdGlvbiBDb21tYW5kSW5wdXRGb3JtKHsgb25FeGVjdXRlLCBkaXJlY3RvcnkgfTogeyBvbkV4ZWN1dGU6IChjb21tYW5kOiBzdHJpbmcpID0+IHZvaWQ7IGRpcmVjdG9yeTogc3RyaW5nIH0pIHtcbiAgY29uc3QgeyBwb3AgfSA9IHVzZU5hdmlnYXRpb24oKTtcbiAgY29uc3QgW2NvbW1hbmRFcnJvciwgc2V0Q29tbWFuZEVycm9yXSA9IHVzZVN0YXRlPHN0cmluZyB8IHVuZGVmaW5lZD4oKTtcblxuICBjb25zdCBjb21tb25Db21tYW5kcyA9IFtcbiAgICBcImxzIC1sYVwiLFxuICAgIFwicHdkXCIsXG4gICAgXCJnaXQgc3RhdHVzXCIsXG4gICAgXCJnaXQgbG9nIC0tb25lbGluZVwiLFxuICAgIFwibnBtIGluc3RhbGxcIixcbiAgICBcIm5wbSBydW4gYnVpbGRcIixcbiAgICBcIm5wbSB0ZXN0XCIsXG4gICAgXCJkb2NrZXIgcHNcIixcbiAgICBcInBzIGF1eFwiLFxuICAgIFwiZGYgLWhcIixcbiAgXTtcblxuICBmdW5jdGlvbiBoYW5kbGVTdWJtaXQodmFsdWVzOiB7IGNvbW1hbmQ6IHN0cmluZyB9KSB7XG4gICAgaWYgKCF2YWx1ZXMuY29tbWFuZC50cmltKCkpIHtcbiAgICAgIHNldENvbW1hbmRFcnJvcihcIlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVx1MzA5Mlx1NTE2NVx1NTI5Qlx1MzA1N1x1MzA2Nlx1MzA0Rlx1MzA2MFx1MzA1NVx1MzA0NFwiKTtcbiAgICAgIHJldHVybjtcbiAgICB9XG5cbiAgICBvbkV4ZWN1dGUodmFsdWVzLmNvbW1hbmQudHJpbSgpKTtcbiAgICBwb3AoKTtcbiAgfVxuXG4gIHJldHVybiAoXG4gICAgPEZvcm1cbiAgICAgIGFjdGlvbnM9e1xuICAgICAgICA8QWN0aW9uUGFuZWw+XG4gICAgICAgICAgPEFjdGlvbi5TdWJtaXRGb3JtIHRpdGxlPVwiXHU1QjlGXHU4ODRDXCIgb25TdWJtaXQ9e2hhbmRsZVN1Ym1pdH0gLz5cbiAgICAgICAgICA8QWN0aW9uIHRpdGxlPVwiXHUzMEFEXHUzMEUzXHUzMEYzXHUzMEJCXHUzMEVCXCIgb25BY3Rpb249e3BvcH0gLz5cbiAgICAgICAgPC9BY3Rpb25QYW5lbD5cbiAgICAgIH1cbiAgICA+XG4gICAgICA8Rm9ybS5EZXNjcmlwdGlvbiB0ZXh0PXtgXHU1QjlGXHU4ODRDXHUzMEM3XHUzMEEzXHUzMEVDXHUzMEFGXHUzMEM4XHUzMEVBOiAke2RpcmVjdG9yeX1gfSAvPlxuICAgICAgPEZvcm0uVGV4dEZpZWxkXG4gICAgICAgIGlkPVwiY29tbWFuZFwiXG4gICAgICAgIHRpdGxlPVwiXHUzMEIzXHUzMERFXHUzMEYzXHUzMEM5XCJcbiAgICAgICAgcGxhY2Vob2xkZXI9XCJcdTVCOUZcdTg4NENcdTMwNTlcdTMwOEJcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTUxNjVcdTUyOUJcdTMwNTdcdTMwNjZcdTMwNEZcdTMwNjBcdTMwNTVcdTMwNDRcIlxuICAgICAgICBlcnJvcj17Y29tbWFuZEVycm9yfVxuICAgICAgICBvbkNoYW5nZT17KCkgPT4gc2V0Q29tbWFuZEVycm9yKHVuZGVmaW5lZCl9XG4gICAgICAvPlxuICAgICAgPEZvcm0uU2VwYXJhdG9yIC8+XG4gICAgICA8Rm9ybS5EZXNjcmlwdGlvbiB0aXRsZT1cIlx1MzA4OFx1MzA0Rlx1NEY3Rlx1MzA0Nlx1MzBCM1x1MzBERVx1MzBGM1x1MzBDOVwiIHRleHQ9XCJcdTRFRTVcdTRFMEJcdTMwNkVcdTMwQjNcdTMwREVcdTMwRjNcdTMwQzlcdTMwOTJcdTUzQzJcdTgwMDNcdTMwNkJcdTMwNTdcdTMwNjZcdTMwNEZcdTMwNjBcdTMwNTVcdTMwNDQ6XCIgLz5cbiAgICAgIHtjb21tb25Db21tYW5kcy5tYXAoKGNtZCwgaW5kZXgpID0+IChcbiAgICAgICAgPEZvcm0uRGVzY3JpcHRpb24ga2V5PXtpbmRleH0gdGV4dD17YFx1MjAyMiAke2NtZH1gfSAvPlxuICAgICAgKSl9XG4gICAgPC9Gb3JtPlxuICApO1xufVxuIl0sCiAgIm1hcHBpbmdzIjogIjs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUEsaUJBY087QUFDUCwyQkFBcUI7QUFDckIsZ0JBQXdCO0FBQ3hCLG1CQUFvQztBQUNwQyxrQkFBMEI7QUFpSmY7QUEvSVgsSUFBTSxnQkFBWSx1QkFBVSx5QkFBSTtBQWtDaEMsSUFBTSxtQkFBOEI7QUFBQSxFQUNsQztBQUFBLElBQ0UsSUFBSTtBQUFBLElBQ0osTUFBTTtBQUFBLElBQ04sU0FBUztBQUFBLElBQ1QsYUFBYTtBQUFBLElBQ2IsZUFBVyxtQkFBUTtBQUFBLEVBQ3JCO0FBQUEsRUFDQTtBQUFBLElBQ0UsSUFBSTtBQUFBLElBQ0osTUFBTTtBQUFBLElB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
