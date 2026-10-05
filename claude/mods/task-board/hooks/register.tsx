import { atom, read, update } from 'claude-code'
import type { Register } from 'claude-code'

import type { Board, BoardList, BoardTask } from '../types'

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

const word = (status: string) => status.split(/[\s(:：]/)[0] ?? ''
const kindOf = (t: BoardTask) => {
  const w = word(t.status)
  if (w === 'merged' || w === 'done') return 'done'
  if (w === 'implementing' || w === 'fixing' || w === 'running') return 'active'
  return 'open'
}
// 状態の補足(waiting (T3後) の括弧など)だけを薄字で添える
const note = (t: BoardTask) => {
  const rest = t.status.slice(word(t.status).length).replace(/^[:：]\s*/, '').trim()
  return rest === '' ? '' : ` ${rest}`
}

const parseList = (name: string, text: string): BoardList | undefined => {
  try {
    const json = JSON.parse(text)
    if (!Array.isArray(json.tasks)) return undefined
    const tasks: BoardTask[] = json.tasks.map((t: any) => ({
      id: String(t.id ?? '?'),
      title: String(t.title ?? ''),
      status: String(t.status ?? ''),
      ws: t.ws == null ? null : String(t.ws),
    }))
    return { name, updatedAt: String(json.updated_at ?? ''), tasks }
  } catch {
    return undefined
  }
}

const allTasks = (lists: BoardList[]) => lists.flatMap(l => l.tasks)

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

const Row = ({ Box, Text, t }: any) => {
  const kind = kindOf(t)
  return (
    <Box key={t.id}>
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
      {kind !== 'done' && (note(t) !== '' || t.ws !== null) ? (
        <Box flexShrink={0}>
          <Text color={SUB}>
            {note(t)}
            {t.ws !== null ? ` @${t.ws}` : ''}
          </Text>
        </Box>
      ) : null}
    </Box>
  )
}

// 手書きの一覧と衝突しないよう、サイドカーは専用ファイルだけを書く
const SIDECAR_FILE = '.claude/tasks/auto.json'
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

const registerSidecar = (on: Parameters<Register>[0]) => {
  let lastRunAt = 0
  let isRunning = false

  on('turn.complete', async ($, e, next) => {
    // 本体の回答が終わったターンだけ。サブエージェントのターンや中断は見ない
    if (e.agentId !== undefined || e.reason !== 'answer' || e.answer.trim() === '') return next(e)
    if (isRunning || Date.now() - lastRunAt < SIDECAR_MIN_INTERVAL_MS) return next(e)

    // .claude/tasks/ があるプロジェクトだけ。全プロジェクトにファイルを作らない
    const hasDir = await $.fs.list(DIR).then(() => true, () => false)
    if (!hasDir) return next(e)

    isRunning = true
    lastRunAt = Date.now()
    const run = async () => {
      const raw = await $.fs.read(SIDECAR_FILE).catch(() => undefined)
      let current: Task[] = []
      try {
        const json = raw === undefined ? undefined : JSON.parse(raw)
        if (Array.isArray(json?.tasks)) current = json.tasks
      } catch {
        // 壊れたファイルは上書きしない
        return
      }
      const reply = await $.model.fork({ prompt: PROMPT(JSON.stringify(current), e.answer) })
      if (!reply.isAnswered) return
      const tasks = parseReply(reply.text, current)
      if (tasks === undefined) return
      const updated_at = new Date().toISOString()
      await $.fs.write(SIDECAR_FILE, JSON.stringify({ updated_at, tasks }, null, 2) + '\n')
    }
    // ターンの終了を待たせない。失敗は黙って捨てる(次のターンでやり直す)
    run()
      .catch(() => {})
      .finally(() => {
        isRunning = false
      })

    return next(e)
  })
}

export const register: Register = on => {
  registerSidecar(on)
  let isPaneOpen = false

  on('session.start', async ($, e, next) => {
    let lastKey: string | undefined

    const poll = async () => {
      let board_: Board
      let key: string
      try {
        const entries = (await $.fs.list(DIR))
          .filter(f => f.kind === 'file' && f.name.endsWith('.json'))
          .sort((a, b) => a.name.localeCompare(b.name))
        const lists: BoardList[] = []
        const brokenFiles: string[] = []
        const texts: string[] = []
        for (const f of entries) {
          const text = await $.fs.read(`${DIR}/${f.name}`).catch(() => undefined)
          texts.push(`${f.name}\0${text}`)
          const list = text === undefined ? undefined : parseList(f.name.replace(/\.json$/, ''), text)
          if (list === undefined) brokenFiles.push(f.name)
          else lists.push(list)
        }
        key = texts.join('\u0001')
        board_ = entries.length === 0 ? { kind: 'none' } : { kind: 'ok', lists, brokenFiles }
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
    await poll()
    $.clock.every(POLL_MS, poll)

    return next(e)
  })

  on('command.run', { command: 'board' }, async $ => {
    if (isPaneOpen) {
      await $.ui.close({ id: PANE })
      isPaneOpen = false
    } else {
      await $.ui.open({ id: PANE, title: 'tasks', focus: true, closeOnEscape: true })
      isPaneOpen = true
    }
    return {}
  })

  // 帯: 標準のタスク一覧と同じ畳み方(未完了の上位だけ + `… +N pending`)
  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    const b = await read($, board)
    if (e.props.hasSurvey || b.kind !== 'ok' || !Array.isArray(b.lists)) return next(e)

    const { Box, Text } = $.ui.resolve(e)
    const tasks = allTasks(b.lists)
    if (tasks.length === 0 && b.brokenFiles.length === 0) return next(e)

    const live = [...tasks.filter(t => kindOf(t) === 'active'), ...tasks.filter(t => kindOf(t) === 'open')]
    const rows = live.slice(0, COLLAPSED_ROWS)
    const omitted = live.length - rows.length

    return (
      <Box flexDirection="column">
        <Header Text={Text} tasks={tasks} hint="  /board で全件" />
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
    const { Box, Text } = $.ui.resolve(e)

    if (b.kind !== 'ok' || !Array.isArray(b.lists)) {
      return <Text color={SUB}>{`  ${DIR}/*.json がありません`}</Text>
    }

    const order = (ts: BoardTask[]) => [
      ...ts.filter(t => kindOf(t) === 'done'),
      ...ts.filter(t => kindOf(t) === 'active'),
      ...ts.filter(t => kindOf(t) === 'open'),
    ]
    const latest = b.lists.map(l => l.updatedAt).sort().pop() ?? ''

    return (
      <Box flexDirection="column">
        <Header Text={Text} tasks={allTasks(b.lists)} hint="" />
        {b.lists.map(l => (
          <Box key={l.name} flexDirection="column">
            {b.lists.length > 1 && <Text color={SUB} bold>{`  ${l.name}`}</Text>}
            {order(l.tasks).map(t => (
              <Row Box={Box} Text={Text} t={t} />
            ))}
          </Box>
        ))}
        {b.brokenFiles.length > 0 && (
          <Text color="red">{`  読めないファイル: ${b.brokenFiles.join(', ')}`}</Text>
        )}
        <Text color={SUB}>{`  ${latest.slice(0, 16).replace('T', ' ')} 更新 · Esc で閉じる`}</Text>
      </Box>
    )
  })
}
