// auto: 会話から推定した status を重ねて表示している印。prev は手書き JSON の status
export type BoardTask = {
  id: string
  title: string
  status: string
  ws: string | null
  auto?: { prev: string; basis: string; at: string; list: string }
}
export type BoardList = { name: string; updatedAt: string; session: string; tasks: BoardTask[] }
// 手書き一覧への変更提案。status が空なら取り消し(fs に削除 API が無いため)
export type Overlay = { list: string; id: string; status: string; prev: string; basis: string; at: string }

export type Board =
  | { kind: 'none' }
  | { kind: 'ok'; lists: BoardList[]; overlays: Overlay[]; brokenFiles: string[]; sessionId: string }

declare module 'claude-code' {
  interface PluginState {
    'task-board': { board: Board }
  }
}
