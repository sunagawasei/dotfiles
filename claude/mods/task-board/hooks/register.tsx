import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { Board, BoardList, BoardTask, Overlay } from '../types'

const board = atom({ plugin: 'task-board', key: 'board' } as const, { kind: 'none' } as Board)

const PANE = 'task-board'
// セッションの作業ディレクトリ(プロジェクト)からの相対パス。1ファイル=1つのタスクリスト
const DIR = '.claude/tasks'
const POLL_MS = 3000
const COLLAPSED_ROWS = 5
// 標準のタスク一覧を ANSI で実測した色(2026-10-04、ghost-visor テーマ)
const SUB = '#aba4c4'
const ACTIVE = '#58caf8'
const DONE = '#76d6c4'

const word = (status: string) => status.split(/[\s(（:：]/)[0] ?? ''
export const kindOf = (t: BoardTask) => {
  const w = word(t.status)
  if (w === 'merged' || w === 'done') return 'done'
  if (w === 'implementing' || w === 'fixing' || w === 'running') return 'active'
  return 'open'
}
// 状態の補足(waiting (T3後) の「T3後」など)。全体を囲む括弧は外す
export const note = (t: BoardTask) => {
  const rest = t.status.slice(word(t.status).length).replace(/^[:：]\s*/, '').trim()
  return rest.replace(/^[(（](.*)[)）]$/, '$1').trim()
}

// updated_at が無い一覧は、ファイルの mtime を更新時刻にする(無いと /board 下部の時刻が他の一覧の値のまま残る)
const parseList = (name: string, text: string, mtimeMs = 0): BoardList | undefined => {
  try {
    const json = JSON.parse(text)
    if (!Array.isArray(json.tasks)) return undefined
    const tasks: BoardTask[] = json.tasks.map((t: any) => ({
      id: String(t.id ?? '?'),
      title: String(t.title ?? ''),
      status: String(t.status ?? ''),
      ws: t.ws == null ? null : String(t.ws),
    }))
    const updatedAt = json.updated_at == null ? (mtimeMs > 0 ? new Date(mtimeMs).toISOString() : '') : String(json.updated_at)
    return { name, updatedAt, session: String(json.session ?? ''), tasks }
  } catch {
    return undefined
  }
}

// keybindings.json で TOGGLE_ACTION に割り当てたキーが、帯・Pane の Button 経由でここへ届く
const TOGGLE_ACTION = 'app:toggleDiffNoiseFilter'
let isPaneOpen = false

async function togglePane($: EngineInterface) {
  if (isPaneOpen) {
    await $.ui.close({ id: PANE })
    isPaneOpen = false
  } else {
    await $.ui.open({ id: PANE, title: 'tasks', focus: true, closeOnEscape: true })
    isPaneOpen = true
  }
}

const allTasks = (lists: BoardList[]) => lists.flatMap(l => l.tasks)

// サイドカーが会話から書く一覧(auto.<セッションID先頭8桁>.json)。手書きの一覧とは別に扱う
export const isAutoList = (l: { name: string }) => l.name === 'auto' || /^auto\.[0-9a-f]{8}$/.test(l.name)

// 現在のセッションIDの先頭8桁を session に含む一覧と、session を持たない手書きの一覧(全セッション共通)だけを出す。
// 別のセッションの一覧は出さない。session を持たない auto 系の一覧は、どのセッションのものか分からないので出さない
export const scopeLists = (lists: BoardList[], sessionId: string) => {
  const short = sessionId.slice(0, 8)
  const mine = short === '' ? [] : lists.filter(l => l.session.includes(short))
  const shared = lists.filter(l => l.session === '' && !isAutoList(l))
  return { lists: [...mine, ...shared.filter(l => !mine.includes(l))], scoped: mine.length > 0 }
}

// 現在のセッションの手書き一覧があれば、それだけを出す(サイドカーの一覧は出さない)。無ければサイドカーの一覧
const scopeAll = (lists: BoardList[], sessionId: string) => {
  const hand = scopeLists(lists.filter(l => !isAutoList(l)), sessionId)
  return hand.lists.length > 0 ? hand : scopeLists(lists.filter(isAutoList), sessionId)
}

// 会話から推定した status を手書き一覧へ重ねる。手書きの status が提案時(prev)から変わっていれば、人の修正を優先して無視する
export const applyOverlays = (lists: BoardList[], overlays: Overlay[]): BoardList[] =>
  lists.map(l => ({
    ...l,
    tasks: l.tasks.map(t => {
      const o = overlays.find(x => x.list === l.name && x.id === t.id)
      if (o === undefined || o.status === '' || o.prev !== t.status) return t
      return { ...t, status: o.status, auto: { prev: t.status, basis: o.basis, at: o.at, list: l.name } }
    }),
  }))

export const viewOf = (lists: BoardList[], overlays: Overlay[], sessionId: string) => {
  const scope = scopeAll(lists, sessionId)
  return { lists: applyOverlays(scope.lists, overlays), scoped: scope.scoped }
}

// 表示中の一覧のうち最も新しい更新時刻。書式が混在する(+0900 と Z)ので文字列でなく時刻で比べる
export const latestUpdate = (lists: BoardList[]) => {
  const ts = lists.map(l => ({ at: l.updatedAt, ms: Date.parse(l.updatedAt) })).filter(x => !Number.isNaN(x.ms))
  return ts.sort((a, b) => a.ms - b.ms).pop()?.at ?? ''
}

export const boardKey = (sessionId: string, listTexts: string[], overlayTexts: string[]) =>
  `${sessionId}\u0002${listTexts.join('\u0001')}\u0003${overlayTexts.join('\u0001')}`

const OVERLAY_DIR = '.claude/task-board/overlay'
// 53bit の文字列ハッシュ(cyrb53)。一覧名と id はファイル本文に持つので、名前は短く一意に近ければよい
const hash53 = (s: string) => {
  let h1 = 0xdeadbeef
  let h2 = 0x41c6ce57
  for (let i = 0; i < s.length; i++) {
    const c = s.charCodeAt(i)
    h1 = Math.imul(h1 ^ c, 2654435761)
    h2 = Math.imul(h2 ^ c, 1597334677)
  }
  h1 = Math.imul(h1 ^ (h1 >>> 16), 2246822507) ^ Math.imul(h2 ^ (h2 >>> 13), 3266489909)
  h2 = Math.imul(h2 ^ (h2 >>> 16), 2246822507) ^ Math.imul(h1 ^ (h1 >>> 13), 3266489909)
  return (4294967296 * (2097151 & h2) + (h1 >>> 0)).toString(16).padStart(14, '0')
}
// 1エントリ=1ファイル。複数セッションが別々のタスクを同時に保存しても互いを消さない。
// 一覧名の連結だと `a__`+`b` と `a`+`__b` が衝突し、日本語名は符号化でファイル名の長さ上限を超えるためハッシュにする
export const overlayPath = (list: string, id: string) => `${OVERLAY_DIR}/${hash53(`${list}\0${id}`)}.json`

const parseOverlay = (text: string): Overlay | undefined => {
  try {
    const j = JSON.parse(text)
    const keys = ['list', 'id', 'status', 'prev', 'basis', 'at'] as const
    if (keys.some(k => typeof j[k] !== 'string')) return undefined
    return { list: j.list, id: j.id, status: j.status, prev: j.prev, basis: j.basis, at: j.at }
  } catch {
    return undefined
  }
}

const KNOWN_WORDS = ['implementing', 'fixing', 'running', 'merged', 'done', 'waiting', 'queued']
const MAX_CHANGES = 10
const MAX_BASIS = 200
const squash = (s: string) => s.replace(/\s+/g, '')

// id が別の id の一部でなく単独で現れること。英数字・`_` に続く/続かれる場合と、`T5-3` `T5.2` のように `-` `.` で英数字へつながる場合は別の id とみなす
// (T5 は「T53」「T5_3」「T5-3」に一致せず、「T5を」「T5-完了」には一致する)
export const mentionsId = (text: string, id: string) => {
  const esc = id.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return new RegExp(`(?<![A-Za-z0-9_]|[A-Za-z0-9][-.])${esc}(?![A-Za-z0-9_]|[-.][A-Za-z0-9])`, 'i').test(text)
}

export type Proposal ={ list: string; id: string; status: string; basis: string }

// fork の返答を検証する。LLM の返答は信用せず、実在する id・既知の語・直前の回答に実在する引用だけを通す
export const parseChanges = (text: string, tasks: { list: string; id: string; shown: string }[], answer: string): Proposal[] => {
  const body = text.trim().replace(/^```(?:json)?\s*/, '').replace(/\s*```$/, '')
  if (body === '' || /^none\b/i.test(body)) return []
  try {
    const json = JSON.parse(body)
    if (!Array.isArray(json.changes) || json.changes.length > MAX_CHANGES) return []
    const seen = new Set<string>()
    const out: Proposal[] = []
    for (const c of json.changes) {
      const list = String(c?.list ?? '')
      const id = String(c?.id ?? '')
      const status = String(c?.status ?? '').trim()
      const basis = String(c?.basis ?? '').trim()
      const task = tasks.find(t => t.list === list && t.id === id)
      if (task === undefined || seen.has(`${list}\0${id}`)) continue
      if (status === '' || !KNOWN_WORDS.includes(word(status))) continue
      if (basis === '' || basis.length > MAX_BASIS || !squash(answer).includes(squash(basis))) continue
      // 別タスクの完了文を根拠にした提案を避けるため、引用が対象の id を名指ししていること
      if (!mentionsId(basis, id)) continue
      if (status === task.shown) continue
      seen.add(`${list}\0${id}`)
      out.push({ list, id, status, basis })
    }
    return out
  } catch {
    return []
  }
}

// 提案を保存用のエントリにする。人手の status と同じ値へ戻す提案は、取り消し(status を空にする)として保存する
export const toOverlay = (p: Proposal, human: string, at: string): Overlay => ({
  list: p.list,
  id: p.id,
  status: p.status === human ? '' : p.status,
  prev: human,
  basis: p.basis,
  at,
})

// コマンド末尾の `# task:T2` や `# task:T2,T3` で名指しされた id。シェルのコメントなのでコミットメッセージには残らない
const TASK_MARK = /#\s*task:\s*([A-Za-z0-9_.-]+(?:\s*,\s*[A-Za-z0-9_.-]+)*)/g
const MAX_MARKED = 10

// 成功した commit / PR merge で完了にするタスクと status。LLM は使わず、タスクの名指しはコマンドの印だけで決める。
// tool.call の結果に終了コードが無いので、成功は出力(`[main abc1234] …` / merged)で見る
export const milestoneOf = (command: string, text: string): { ids: string[]; status: string; basis: string } | undefined => {
  const ids = [...new Set([...command.matchAll(TASK_MARK)].flatMap(m => m[1]!.split(',').map(s => s.trim())))].slice(0, MAX_MARKED)
  if (ids.length === 0) return undefined
  if (/\bgit\s+(-C\s+\S+\s+)?commit\b/.test(command)) {
    const sha = text.match(/^\[[^\]\n]+\s([0-9a-f]{7,40})\]/m)?.[1]
    if (sha !== undefined) return { ids, status: `done (commit ${sha})`, basis: `git commit ${sha}` }
  }
  if (/\bgh\s+pr\s+merge\b/.test(command) && /merged/i.test(text)) return { ids, status: 'merged', basis: 'gh pr merge' }
  return undefined
}

// 名指しされた行の status を差し替える。既定では完了済みの行(補足つきの done など)を上書きしない。明示の指定(task_set)だけ上書きを許す。
// 同じ status への差し替えは変更とみなさない。変更が無ければ undefined
export const markTasks = <T extends { id: string; status: string }>(
  tasks: T[],
  ids: string[],
  status: string,
  overwriteDone = false,
): T[] | undefined => {
  let changed = false
  const out = tasks.map(t => {
    if (!ids.includes(t.id) || t.status === status) return t
    if (!overwriteDone && ['done', 'merged'].includes(word(t.status))) return t
    changed = true
    return { ...t, status }
  })
  return changed ? out : undefined
}

const Header = ({ Text, tasks, hint }: any) => {
  const done = tasks.filter((t: BoardTask) => kindOf(t) === 'done').length
  const active = tasks.filter((t: BoardTask) => kindOf(t) === 'active').length
  const open = tasks.length - done - active
  return (
    <Text color={SUB} wrap="truncate-end">
      {'  '}
      <Text bold>{tasks.length}</Text> tasks (<Text bold>{done}</Text> done,{' '}
      <Text bold>{active}</Text> in progress, <Text bold>{open}</Text> open){hint}
    </Text>
  )
}

// 会話から推定した行の印。手書き JSON の値を添えて見比べられるようにする
export const autoMark = (t: BoardTask) =>
  t.auto === undefined ? '' : `auto(推定) JSON: ${t.auto.prev.length > 24 ? `${t.auto.prev.slice(0, 24)}…` : t.auto.prev}`

const stamp = (at: string) => {
  const d = new Date(at)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

const hhmm = (at: string) => {
  const d = new Date(at)
  return Number.isNaN(d.getTime()) ? at.slice(11, 16) : `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

// Pane 下部の「自動判定」行。根拠は直前の回答からの引用であって、検証済みの事実ではない
export const autoNote = (t: BoardTask) =>
  t.auto === undefined ? '' : `  auto(推定) ${hhmm(t.auto.at)} ${t.auto.list}/${t.id} ← 「${t.auto.basis.length > 60 ? `${t.auto.basis.slice(0, 60)}…` : t.auto.basis}」`

// 1行目はタイトルだけ。補足・ws・auto の印は、Pane(withDetail)だけ2行目へインデントして出す。帯はタイトルのみ
const Row = ({ Box, Text, t, withDetail = false }: any) => {
  const kind = kindOf(t)
  const detail = [
    kind !== 'done' ? note(t) : '',
    kind !== 'done' && t.ws !== null ? `@${t.ws}` : '',
    autoMark(t),
  ]
    .filter(x => x !== '')
    .join('  ')
  return (
    <Box key={t.id} flexDirection="column">
      <Box>
        <Box flexShrink={0}>
          <Text color={kind === 'active' ? ACTIVE : kind === 'done' ? DONE : undefined}>
            {kind === 'done' ? '  ✔ ' : kind === 'active' ? '  ◼ ' : '  ◻ '}
          </Text>
        </Box>
        <Text
          bold={kind === 'active'}
          color={kind === 'done' ? SUB : undefined}
          strikethrough={kind === 'done'}
          wrap="truncate-end"
        >
          {t.id}: {t.title}
        </Text>
      </Box>
      {withDetail && detail !== '' ? (
        <Box paddingLeft={6}>
          <Text color={SUB}>{detail}</Text>
        </Box>
      ) : null}
    </Box>
  )
}

const SIDECAR_MIN_INTERVAL_MS = 30_000
const MAX_TASKS = 50

type Task = { id: string; title: string; status: string; ws: string | null }

const PROMPT = (current: string, answer: string) => `あなたはタスク一覧の更新係です。ここまでの会話と直前の回答から、作業タスクの一覧を最新に保ちます。

現在の一覧:
${current}

直前の回答:
${answer}

ルール:
- 一覧を変える必要が無ければ、\`none\` の1語だけを返す。
- 変える場合は、更新後の一覧全体を次のJSONだけで返す。説明やコードフェンスは付けない。
  {"tasks":[{"id":"T1","title":"短いタイトル","status":"implementing","ws":null}]}
- status の先頭の語は implementing / fixing / running(進行中)、merged / done(完了)、waiting / queued(未着手)のどれか。
- 既存の id は変えない。新しいタスクは次の連番(T<n>)にする。
- 会話で明示された作業だけを載せる。推測で足さない。完了は完了を確認できたものだけ。
- 次の場合は、変更があっても none を返す: 雑談、質問への回答だけで作業が発生していない。`

export const parseReply = (text: string, current: Task[]): Task[] | undefined => {
  const body = text.trim().replace(/^```(?:json)?\s*/, '').replace(/\s*```$/, '')
  if (body === '' || /^none\b/i.test(body)) return undefined
  try {
    const json = JSON.parse(body)
    if (!Array.isArray(json.tasks) || json.tasks.length > MAX_TASKS) return undefined
    const tasks: Task[] = json.tasks.map((t: any) => ({
      id: String(t.id ?? ''),
      title: String(t.title ?? ''),
      status: String(t.status ?? ''),
      ws: t.ws == null ? null : String(t.ws),
    }))
    if (tasks.some(t => t.id === '' || t.title === '' || t.status === '')) return undefined
    if (new Set(tasks.map(t => t.id)).size !== tasks.length) return undefined
    return JSON.stringify(tasks) === JSON.stringify(current) ? undefined : tasks
  } catch {
    return undefined
  }
}

const OVERLAY_PROMPT = (tasks: string, answer: string) => `あなたはタスク一覧の更新係です。手書きのタスク一覧の status が、会話の実態とずれていないか確かめます。

現在のタスク(status は手書きの値、shown は画面に出ている値):
${tasks}

未処理の回答(古い順):
${answer}

ルール:
- 変える必要が無ければ、\`none\` の1語だけを返す。
- 変える場合は、次のJSONだけを返す。説明やコードフェンスは付けない。
  {"changes":[{"list":"一覧名","id":"T1","status":"merged","basis":"回答からの逐語引用"}]}
- basis に選ぶ文は、変えるタスクの id(例 T1)を名指ししているものにする。別のタスクの文を根拠にしない。
- basis は「回答」にある文を、一字も変えずに引用する(${MAX_BASIS}字以内)。回答に無い根拠は使わない。
- status の先頭の語は implementing / fixing / running(進行中)、merged / done(完了)、waiting / queued(未着手)のどれか。
- 完了にするのは、回答で完了が明示されたときだけ。「完了にする予定」「これからやる」は完了にしない。worker の自己報告だけでは完了にしない。
- implementing にするのは、着手や起動が回答で明示されたときだけ。
- shown が誤りだと回答で訂正されたら、status に手書きの status をそのまま入れて戻す。
- 推測で変えない。雑談や、質問への回答だけで作業が発生していなければ none を返す。
- title と ws は変えない。`

type Loaded = Awaited<ReturnType<typeof loadBoard>>

export const runOverlay = async ($: EngineInterface, data: Loaded, hand: BoardList[], answer: string, root: string) => {
  const applied = applyOverlays(hand, data.overlays)
  const tasks = hand.flatMap((l, i) =>
    l.tasks.map((t, j) => ({ list: l.name, id: t.id, title: t.title, status: t.status, shown: applied[i]!.tasks[j]!.status })),
  )
  if (tasks.length === 0) return
  const reply = await $.model.fork({ prompt: OVERLAY_PROMPT(JSON.stringify(tasks), answer) })
  if (!reply.isAnswered) return
  const at = new Date().toISOString()
  for (const p of parseChanges(reply.text, tasks, answer)) {
    const human = tasks.find(t => t.list === p.list && t.id === p.id)!.status
    const rel = overlayPath(p.list, p.id)
    const path = `${root}/${rel}`
    // 提案は fork に渡した時点の overlay を前提に作られている。
    // 書く直前の overlay がその時点と違えば、task_set などが書いた新しい値を守るため提案を捨てる
    const prefix = `${rel.slice(rel.lastIndexOf('/') + 1)}\0`
    const seen = data.overlayTexts.find(x => x.startsWith(prefix))?.slice(prefix.length)
    await serialized(async () => {
      if ((await $.fs.read(path).catch(() => undefined)) !== seen) return
      await $.fs.write(path, JSON.stringify(toOverlay(p, human, at), null, 2) + '\n')
    })
  }
}

// 現在のセッションの手書き一覧が無いときは、会話から一覧そのものを作る。セッションごとに別ファイルへ書く
export const autoFile = (short: string) => `${DIR}/auto.${short}.json`

const runLegacy = async ($: EngineInterface, answer: string, root: string, sessionId: string) => {
  const short = sessionId.slice(0, 8)
  if (short === '') return
  const path = `${root}/${autoFile(short)}`
  const raw = await $.fs.read(path).catch(() => undefined)
  let current: Task[] = []
  try {
    const json = raw === undefined ? undefined : JSON.parse(raw)
    if (Array.isArray(json?.tasks)) current = json.tasks
  } catch {
    // 壊れたファイルは上書きしない
    return
  }
  const reply = await $.model.fork({ prompt: PROMPT(JSON.stringify(current), answer) })
  if (!reply.isAnswered) return
  const tasks = parseReply(reply.text, current)
  if (tasks === undefined) return
  const updated_at = new Date().toISOString()
  await serialized(async () => {
    // fork の間に節目フックや task_set が書いた変更を、古い一覧で上書きしない
    if ((await $.fs.read(path).catch(() => undefined)) !== raw) return
    await $.fs.write(path, JSON.stringify({ updated_at, session: `session-${short}`, tasks }, null, 2) + '\n')
  })
}

const MAX_PENDING_ANSWERS = 8
const MAX_ANSWER_CHARS = 8000

export type AnswerScope = { cwd: string; sessionId: string }
export type Pending = AnswerScope & { answers: string[] }

// 間隔内に終わった回答を古い順に溜める。最新の1件だけだと、完了を告げた回答の直後の「ありがとう」で上書きされる。
// cwd か session ID が変わったら(/cd・/clear)別の場所・別の会話の回答なので溜め直す。長い回答は、結論が載りやすい末尾を残す
export const queueAnswer = (p: Pending | undefined, scope: AnswerScope, answer: string): Pending => {
  const isSame = p !== undefined && p.cwd === scope.cwd && p.sessionId === scope.sessionId
  const kept = isSame ? p.answers : []
  const text = answer.length > MAX_ANSWER_CHARS ? `${answer.slice(0, 2000)}…${answer.slice(-6000)}` : answer
  return { cwd: scope.cwd, sessionId: scope.sessionId, answers: [...kept, text].slice(-MAX_PENDING_ANSWERS) }
}

export const joinAnswers = (answers: string[]) => answers.join('\n\n---\n\n')

// 回答の間隔制御の状態。kick が $ を受け取るため、validate の制約でトップレベルに置く
const sidecar = { lastRunAt: 0, isRunning: false, pending: undefined as Pending | undefined }

// 溜めた回答を、間隔が空いていれば処理する。turn.complete と poll(3秒ごと)から呼ぶ。
// 間隔内に終わった最後のターンも、次の poll で処理される(タイマーに頼らないので、拒否や失敗で取り残されない)
const kick = async ($: EngineInterface): Promise<void> => {
  if (sidecar.isRunning || sidecar.pending === undefined) return
  if (Date.now() - sidecar.lastRunAt < SIDECAR_MIN_INTERVAL_MS) return
  const batch = sidecar.pending
  sidecar.pending = undefined
  sidecar.isRunning = true
  sidecar.lastRunAt = Date.now()
  try {
    // 待機中に /cd や /clear で移ったら、溜めた回答は別の repo・別の会話のものなので捨てる。
    // 処理の途中で移っても読み先・書き込み先がずれないよう、読み書きは batch.cwd の絶対パスで行う
    if (batch.cwd === '' || batch.sessionId === '') return
    if ((await $.session.cwd()) !== batch.cwd || (await $.session.id().catch(() => '')) !== batch.sessionId) return
    const answer = joinAnswers(batch.answers)
    const data = await loadBoard($, batch.cwd)
    // 現在のセッションの手書き一覧(session 無しの共通一覧を含む)があれば手書きへの提案、無ければ会話から一覧を作る
    const hand = scopeLists(data.lists.filter(l => !isAutoList(l)), batch.sessionId).lists
    if (hand.length > 0) await runOverlay($, data, hand, answer, batch.cwd)
    else await runLegacy($, answer, batch.cwd, batch.sessionId)
  } catch {
    // 失敗は黙って捨てる(次のターンでやり直す)
  } finally {
    sidecar.isRunning = false
  }
}

// 状態を書く処理(節目フック・task_set・サイドカー)を1本の列に並べる。並列呼び出しでも read-modify-write が交錯せず、結果は呼び出し順になる
let writeQueue: Promise<unknown> = Promise.resolve()
export const serialized = <T,>(fn: () => Promise<T>): Promise<T> => {
  const run = writeQueue.then(fn, fn)
  writeQueue = run.catch(() => {})
  return run
}

export type Ref = { list: string; id: string }
export type StatusChange = { ids: string[]; list?: string; status: string; basis: string; isExplicit: boolean }
// applied: 書いたタスク / known: 対象の一覧にあるタスク / ambiguous: list 無しで id が複数の一覧にあったときの一覧名(このとき何も書かない) / isMulti: 手書き一覧が複数
export type StatusResult = { applied: Ref[]; known: Ref[]; ambiguous: string[]; isMulti: boolean }

// 名指しされたタスクの status を、現在のセッションの一覧へその場で反映する。
// 現在のセッションの手書き一覧があれば overlay(画面だけ重ねる)、無ければ auto 一覧へ書く。
// isExplicit(task_set)なら完了済みの行も上書きし、節目フックは完了済みの行を守る。
// 同じ id が複数の手書き一覧にあり list の指定が無いときは、どれを変えるか決められないので何も書かない
export const applyStatus = async ($: EngineInterface, c: StatusChange): Promise<StatusResult> => {
  const none: StatusResult = { applied: [], known: [], ambiguous: [], isMulti: false }
  const cwd = await $.session.cwd().catch(() => '')
  const sessionId = await $.session.id().catch(() => '')
  if (cwd === '' || sessionId === '') return none
  if (!(await $.fs.list(`${cwd}/${DIR}`).then(() => true, () => false))) return none
  const data = await loadBoard($, cwd)
  const at = new Date().toISOString()
  const hand = scopeLists(data.lists.filter(l => !isAutoList(l)), sessionId).lists
  const applied: Ref[] = []
  if (hand.length > 0) {
    const known = hand.flatMap(l => l.tasks.map(t => ({ list: l.name, id: t.id })))
    const isMulti = hand.length > 1
    const inScope = (l: BoardList) => c.list === undefined || l.name === c.list
    const ambiguous = c.isExplicit && c.ids.length === 1 && c.list === undefined ? hand.filter(l => l.tasks.some(t => t.id === c.ids[0])).map(l => l.name) : []
    if (ambiguous.length > 1) return { applied, known, ambiguous, isMulti }
    const shown = applyOverlays(hand, data.overlays)
    for (const [i, l] of hand.entries()) {
      if (!inScope(l)) continue
      for (const [j, t] of l.tasks.entries()) {
        if (!c.ids.includes(t.id)) continue
        const cur = shown[i]!.tasks[j]!
        if (cur.status === c.status || (!c.isExplicit && kindOf(cur) === 'done')) continue
        const o = toOverlay({ list: l.name, id: t.id, status: c.status, basis: c.basis }, t.status, at)
        await $.fs.write(`${cwd}/${overlayPath(l.name, t.id)}`, JSON.stringify(o, null, 2) + '\n')
        applied.push({ list: l.name, id: t.id })
      }
    }
    return { applied, known, ambiguous: [], isMulti }
  }
  const name = `auto.${sessionId.slice(0, 8)}`
  const path = `${cwd}/${autoFile(sessionId.slice(0, 8))}`
  const raw = await $.fs.read(path).catch(() => undefined)
  if (raw === undefined) return none
  const json = JSON.parse(raw)
  if (!Array.isArray(json.tasks)) return none
  const known = (json.tasks as Task[]).map(t => ({ list: name, id: t.id }))
  // list の指定が auto 一覧と合わないときは書かない(呼び出し側が「見つかりません」と断るので、書くと失敗を返しながら表示が変わる)
  if (c.list !== undefined && c.list !== name) return { applied, known, ambiguous: [], isMulti: false }
  const tasks = markTasks<Task>(json.tasks, c.ids, c.status, c.isExplicit)
  if (tasks === undefined) return { applied, known, ambiguous: [], isMulti: false }
  await $.fs.write(path, JSON.stringify({ ...json, updated_at: at, tasks }, null, 2) + '\n')
  return { applied: known.filter(k => c.ids.includes(k.id)), known, ambiguous: [], isMulti: false }
}

const TOOL_NAME = 'task_set'
const TOOL_FULL = `mcp__task-board__${TOOL_NAME}`
const SETTABLE = KNOWN_WORDS.join(' / ')

// 手書き一覧が複数あるときだけ「一覧名/id」で示す(どの一覧のタスクか分かるように)
export const refLabel = (r: Ref, isMulti: boolean) => (isMulti ? `${r.list}/${r.id}` : r.id)

// task_set の入力を検証する。モデルの入力は信用せず、実在の語と空でない id だけ通す
export const parseSetInput = (e: {
  id?: unknown
  status?: unknown
  note?: unknown
  list?: unknown
}): { id: string; status: string; list?: string } | string => {
  const id = typeof e.id === 'string' ? e.id.trim() : ''
  const w = typeof e.status === 'string' ? e.status.trim() : ''
  if (id === '') return 'id が空です。一覧のタスク id(T1 など)を渡してください'
  if (!KNOWN_WORDS.includes(w)) return `status は ${SETTABLE} のどれかです`
  const note = typeof e.note === 'string' ? e.note.replace(/\s+/g, ' ').trim().slice(0, 100) : ''
  const list = typeof e.list === 'string' && e.list.trim() !== '' ? e.list.trim() : undefined
  return { id, status: note === '' ? w : `${w} (${note})`, ...(list === undefined ? {} : { list }) }
}

// task_set を実行して、モデルへ返す結果(result か deny)を作る
export const runTaskSet = async ($: EngineInterface, input: { id: string; status: string; list?: string }) => {
  const r = await serialized(() =>
    applyStatus($, { ids: [input.id], ...(input.list === undefined ? {} : { list: input.list }), status: input.status, basis: TOOL_NAME, isExplicit: true }),
  )
  if (r.known.length === 0) return { deny: 'このセッションの一覧がありません。.claude/tasks/ に一覧を作ってから呼んでください' }
  if (r.ambiguous.length > 1) return { deny: `id ${input.id} は複数の一覧(${r.ambiguous.join(', ')})にあります。list にどれかを指定してください` }
  const hit = r.known.filter(k => k.id === input.id && (input.list === undefined || k.list === input.list))
  if (hit.length === 0) {
    const all = r.known.map(k => refLabel(k, r.isMulti)).join(', ')
    return { deny: `${input.list === undefined ? '' : `一覧 ${input.list} の `}id ${input.id} は見つかりません。あるのは ${all} です` }
  }
  const label = refLabel(hit[0]!, r.isMulti)
  return { result: r.applied.length > 0 ? `${label} を ${input.status} にしました` : `${label} は既に ${input.status} です` }
}

// モデルが呼ぶ task_set の登録。スキーマをプロンプトに常駐させる(ToolSearch の奥だとモデルが呼び忘れる)。
// .claude/tasks/ があるとき、未登録なら登録する。同名の再登録は置き換えなので、失敗した回は次の poll でやり直す
const TOOL_SPEC = {
  name: TOOL_NAME,
  description:
    'task-board の一覧のタスクの状態を、その場で更新する。着手・完了・やり直しのたびに呼ぶ。コマンドの成否ではなく自分の判断で呼ぶ。id は一覧のタスク id(T1 など)。完了にするのは実際に終わったときだけ。',
  inputSchema: {
    type: 'object',
    properties: {
      id: { type: 'string', description: '一覧のタスク id(T1 など)' },
      status: { type: 'string', enum: KNOWN_WORDS, description: '進行中は implementing / fixing / running、完了は done / merged、未着手は waiting / queued' },
      note: { type: 'string', description: '状態の補足(任意・100字まで)。例: 調査完了、stgデプロイ待ち' },
      list: { type: 'string', description: '一覧名(任意)。同じ id が複数の一覧にあるときだけ指定する' },
    },
    required: ['id', 'status'],
  },
  isDeferred: false,
}
export const toolState = { isRegistered: false }
export const ensureTool = async ($: EngineInterface, state = toolState): Promise<void> => {
  if (state.isRegistered) return
  if (!(await $.fs.list(DIR).then(() => true, () => false))) return
  await $.tool.register(TOOL_SPEC).then(
    () => {
      state.isRegistered = true
    },
    () => {},
  )
}

const registerSidecar = (on: Parameters<Register>[0]) => {
  // 回答の終了を待たず、commit / PR merge の成功をその場で反映する。結果は変えず、反映は待たない
  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const r = await next(e)
    try {
      const text = 'text' in r && typeof r.text === 'string' ? r.text : ''
      const m = r.isError || r.deny !== undefined || text === '' ? undefined : milestoneOf(e.command, text)
      if (m !== undefined) serialized(() => applyStatus($, { ...m, isExplicit: false })).catch(() => {})
    } catch {
      // 反映の失敗でツールの結果を壊さない
    }
    return r
  })

  // モデルが明示的に呼ぶ状態更新。コマンドの成否ではなく、モデルの判断で着手・完了・やり直しを即座に反映する
  on('tool.call', { tool: TOOL_FULL }, async ($, e) => {
    const input = parseSetInput(e)
    if (typeof input === 'string') return { deny: input }
    try {
      return await runTaskSet($, input)
    } catch {
      return { deny: '一覧の更新に失敗しました。.claude/tasks/ のファイルを直接直してください' }
    }
  })

  on('turn.complete', async ($, e, next) => {
    // 本体の回答が終わったターンだけ。サブエージェントのターンや中断は見ない
    if (e.agentId !== undefined || e.reason !== 'answer' || e.answer.trim() === '') return next(e)

    // .claude/tasks/ があるプロジェクトだけ。全プロジェクトにファイルを作らない
    const hasDir = await $.fs.list(DIR).then(() => true, () => false)
    if (!hasDir) return next(e)

    // ターンの終了を待たせない
    const cwd = await $.session.cwd().catch(() => '')
    const sessionId = await $.session.id().catch(() => '')
    if (cwd === '' || sessionId === '') return next(e)
    sidecar.pending = queueAnswer(sidecar.pending, { cwd, sessionId }, e.answer)
    kick($).catch(() => {})
    return next(e)
  })
}

// root を渡すと、その repo の絶対パスで読む(処理の途中で /cd されても読み先がずれない)
export const loadBoard = async ($: EngineInterface, root = '') => {
  const at = (p: string) => (root === '' ? p : `${root}/${p}`)
  // /clear では session.start が再発火せず ID だけ変わるため、毎回読む
  const sessionId = await $.session.id().catch(() => '')
  const byName = (a: { name: string }, b: { name: string }) => a.name.localeCompare(b.name)
  const entries = (await $.fs.list(at(DIR))).filter(f => f.kind === 'file' && f.name.endsWith('.json')).sort(byName)
  const lists: BoardList[] = []
  const brokenFiles: string[] = []
  const listTexts: string[] = []
  for (const f of entries) {
    const text = await $.fs.read(at(`${DIR}/${f.name}`)).catch(() => undefined)
    listTexts.push(`${f.name}\0${text}`)
    const list = text === undefined ? undefined : parseList(f.name.replace(/\.json$/, ''), text, f.mtimeMs)
    if (list === undefined) brokenFiles.push(f.name)
    else lists.push(list)
  }
  const overlayEntries = await $.fs
    .list(at(OVERLAY_DIR))
    .then(es => es.filter(f => f.kind === 'file' && f.name.endsWith('.json')).sort(byName), () => [])
  const overlays: Overlay[] = []
  const overlayTexts: string[] = []
  for (const f of overlayEntries) {
    const text = await $.fs.read(at(`${OVERLAY_DIR}/${f.name}`)).catch(() => undefined)
    overlayTexts.push(`${f.name}\0${text}`)
    const o = text === undefined ? undefined : parseOverlay(text)
    if (o !== undefined) overlays.push(o)
  }
  return { sessionId, entryCount: entries.length, lists, overlays, brokenFiles, listTexts, overlayTexts }
}

export const register: Register = on => {
  registerSidecar(on)

  on('session.start', async ($, e, next) => {
    let lastKey: string | undefined

    const poll = async () => {
      // 間隔内に終わった最後の回答を、間隔が空いた後に処理する
      kick($).catch(() => {})
      // /cd で .claude/tasks/ のある repo へ移ったときも、数秒で task_set を使えるようにする
      ensureTool($).catch(() => {})
      let board_: Board
      let key: string
      try {
        const d = await loadBoard($)
        key = boardKey(d.sessionId, d.listTexts, d.overlayTexts)
        board_ =
          d.entryCount === 0
            ? { kind: 'none' }
            : { kind: 'ok', lists: d.lists, overlays: d.overlays, brokenFiles: d.brokenFiles, sessionId: d.sessionId }
      } catch {
        key = ''
        board_ = { kind: 'none' }
      }
      if (key === lastKey) return
      lastKey = key
      await update($, board, () => board_)
    }

    await $.command.register({
      name: 'board',
      description: 'task-board の全件(完了を含む)を Pane で開く / 閉じる',
      immediate: true,
    })
    // 初回のモデル要求に task_set を載せるため、最初の登録は待つ(以降は poll が見る)
    await ensureTool($).catch(() => {})
    await poll()
    $.clock.every(POLL_MS, poll)

    return next(e)
  })

  on('command.run', { command: 'board' }, async $ => {
    await togglePane($)
    return {}
  })

  // 帯: 標準のタスク一覧と同じ畳み方(未完了の上位だけ + `… +N pending`)
  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    const b = await read($, board)
    if (e.props.hasSurvey || b.kind !== 'ok' || !Array.isArray(b.lists)) return next(e)

    const { Box, Button, Text } = $.ui.resolve(e)
    const tasks = allTasks(viewOf(b.lists, b.overlays, b.sessionId).lists)
    if (tasks.length === 0 && b.brokenFiles.length === 0) return next(e)

    const live = [...tasks.filter(t => kindOf(t) === 'active'), ...tasks.filter(t => kindOf(t) === 'open')]
    const rows = live.slice(0, COLLAPSED_ROWS)
    const omitted = live.length - rows.length

    return (
      <Box flexDirection="column">
        <Box>
          <Header Text={Text} tasks={tasks} hint="  全件を開く:" />
          <Box flexShrink={0}>
            <Button key="toggle" label=" ctrl+x t" plain dimColor action={TOGGLE_ACTION} onPress={() => togglePane($)} />
          </Box>
        </Box>
        {rows.map(t => (
          <Row Box={Box} Text={Text} t={t} />
        ))}
        {omitted > 0 && <Text color={SUB}>{`  … +${omitted} pending`}</Text>}
        {b.brokenFiles.length > 0 && (
          <Text color="red" wrap="truncate-end">{`  読めないファイル: ${b.brokenFiles.join(', ')}`}</Text>
        )}
      </Box>
    )
  })

  // Pane: 完了を含む全件。高さを超えれば Pane 側でスクロールする
  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e, next) => {
    const b = await read($, board)
    const { Box, Button, Text } = $.ui.resolve(e)

    if (b.kind !== 'ok' || !Array.isArray(b.lists)) {
      return <Text color={SUB}>{`  ${DIR}/*.json がありません`}</Text>
    }

    const order = (ts: BoardTask[]) => [
      ...ts.filter(t => kindOf(t) === 'done'),
      ...ts.filter(t => kindOf(t) === 'active'),
      ...ts.filter(t => kindOf(t) === 'open'),
    ]
    const scope = viewOf(b.lists, b.overlays, b.sessionId)
    const latest = latestUpdate(scope.lists)
    // 会話から推定して重ねた全ての行の、判定時刻と根拠(新しい順)
    const autos = allTasks(scope.lists)
      .filter(t => t.auto !== undefined)
      .sort((a, b) => b.auto!.at.localeCompare(a.auto!.at))

    return (
      <Box flexDirection="column">
        <Box>
          <Header Text={Text} tasks={allTasks(scope.lists)} hint={scope.scoped ? '  · このセッションの一覧だけ表示中' : ''} />
          <Box flexShrink={0}>
            <Button key="toggle" label=" ctrl+x t で閉じる" plain dimColor action={TOGGLE_ACTION} onPress={() => togglePane($)} />
          </Box>
        </Box>
        {scope.lists.map(l => (
          <Box key={l.name} flexDirection="column">
            {scope.lists.length > 1 && <Text color={SUB} bold>{`  ${l.name}`}</Text>}
            {order(l.tasks).map(t => (
              <Row Box={Box} Text={Text} t={t} withDetail />
            ))}
          </Box>
        ))}
        {autos.map(t => (
          <Text key={`auto-${t.auto!.list}-${t.id}`} color={SUB} wrap="truncate-end">{autoNote(t)}</Text>
        ))}
        {b.brokenFiles.length > 0 && (
          <Text color="red">{`  読めないファイル: ${b.brokenFiles.join(', ')}`}</Text>
        )}
        <Text color={SUB}>{`  ${latest === '' ? '' : `${stamp(latest)} 更新 · `}Esc で閉じる`}</Text>
      </Box>
    )
  })
}
