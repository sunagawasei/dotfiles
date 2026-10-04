export type BoardTask = { id: string; title: string; status: string; ws: string | null }
export type BoardList = { name: string; updatedAt: string; tasks: BoardTask[] }

export type Board =
  | { kind: 'none' }
  | { kind: 'ok'; lists: BoardList[]; brokenFiles: string[] }

declare module 'claude-code' {
  interface PluginState {
    'task-board': { board: Board }
  }
}
