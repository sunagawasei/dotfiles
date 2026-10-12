import { test, expect } from 'claude-code/testing'

import {
  applyOverlays,
  applyStatus,
  ensureTool,
  loadBoard,
  runOverlay,
  runTaskSet,
  autoMark,
  autoNote,
  boardKey,
  kindOf,
  markTasks,
  milestoneOf,
  latestUpdate,
  isAutoList,
  joinAnswers,
  mentionsId,
  note,
  overlayPath,
  parseSetInput,
  serialized,
  parseChanges,
  queueAnswer,
  scopeLists,
  toOverlay,
  viewOf,
  autoFile,
} from '../hooks/register'

const t = (status: string) => ({ id: 'T1', title: 'x', status, ws: null })

// [status, 分類, 補足]
const CASES: [string, string, string][] = [
  ['done', 'done', ''],
  ['done: x', 'done', 'x'],
  ['done：x', 'done', 'x'],
  ['done: 該当0件', 'done', '該当0件'],
  ['merged', 'done', ''],
  ['implementing', 'active', ''],
  ['fixing', 'active', ''],
  ['running', 'active', ''],
  ['waiting (T1後)', 'open', 'T1後'],
  ['waiting（T1後）', 'open', 'T1後'],
  ['queued', 'open', ''],
  ['', 'open', ''],
]

for (const [status, kind, rest] of CASES) {
  test(`status「${status}」は ${kind}・補足「${rest}」`, async () => {
    expect(kindOf(t(status))).toBe(kind)
    expect(note(t(status))).toBe(rest)
  })
}

const list = (name: string, session: string) => ({ name, updatedAt: '', session, tasks: [] })
const A = list('a', 'session-0f692c4e')
const B = list('b', 'session-11111111')
const C = list('c', '')

const SID = '0f692c4e-aaaa-bbbb-cccc-dddddddddddd'
const S = (cwd: string, sessionId = 'sess-1') => ({ cwd, sessionId })

test('現在のセッションに一致する一覧と、session の無い手書きの一覧(共通)だけを出す', async () => {
  const r = scopeLists([A, B, C], SID)
  expect(r.scoped).toBe(true)
  expect(r.lists).toEqual([A, C])
})

test('一致する一覧が無ければ、別のセッションの一覧は出さず、共通の一覧だけを出す', async () => {
  const r = scopeLists([B, C], SID)
  expect(r.scoped).toBe(false)
  expect(r.lists).toEqual([C])
})

test('一致する一覧も共通の一覧も無ければ、何も出さない', async () => {
  expect(scopeLists([B], SID).lists).toEqual([])
})

test('セッションIDが空なら、共通の一覧だけを出す', async () => {
  const r = scopeLists([A, C], '')
  expect(r.scoped).toBe(false)
  expect(r.lists).toEqual([C])
})

test('session の無い auto 系の一覧は、どのセッションのものか分からないので出さない', async () => {
  const stale = list('auto', '')
  const mine = list('auto.0f692c4e', 'session-0f692c4e')
  expect(scopeLists([stale], SID).lists).toEqual([])
  expect(scopeLists([stale, mine], SID).lists).toEqual([mine])
})

// --- 会話から推定した status の重ね合わせ ---

const task = (id: string, status: string) => ({ id, title: `title ${id}`, status, ws: null })
const mk = (name: string, session: string, tasks: ReturnType<typeof task>[]) => ({ name, updatedAt: '', session, tasks })
const ov = (list: string, id: string, status: string, prev: string) => ({ list, id, status, prev, basis: 'b', at: '2026-10-05T12:00:00.000Z' })

test('手書きの status が提案時と同じなら、overlay の status を auto つきで表示する', async () => {
  const [l] = applyOverlays([mk('a', '', [task('T1', 'implementing')])], [ov('a', 'T1', 'merged', 'implementing')])
  expect(l!.tasks[0]!.status).toBe('merged')
  expect(l!.tasks[0]!.auto).toEqual({ prev: 'implementing', basis: 'b', at: '2026-10-05T12:00:00.000Z', list: 'a' })
})

test('人が status を書き換えたら overlay を無視する', async () => {
  const [l] = applyOverlays([mk('a', '', [task('T1', 'queued')])], [ov('a', 'T1', 'merged', 'implementing')])
  expect(l!.tasks[0]!.status).toBe('queued')
  expect(l!.tasks[0]!.auto).toBeUndefined()
})

test('別の一覧の同じ id には overlay を適用しない', async () => {
  const [a, b] = applyOverlays(
    [mk('a', '', [task('T1', 'implementing')]), mk('b', '', [task('T1', 'implementing')])],
    [ov('a', 'T1', 'merged', 'implementing')],
  )
  expect(a!.tasks[0]!.status).toBe('merged')
  expect(b!.tasks[0]!.status).toBe('implementing')
})

test('status が空の overlay(取り消し)は適用しない', async () => {
  const [l] = applyOverlays([mk('a', '', [task('T1', 'implementing')])], [ov('a', 'T1', '', 'implementing')])
  expect(l!.tasks[0]!.status).toBe('implementing')
})

test('現在のセッションの手書き一覧があれば、サイドカーの一覧は出さない', async () => {
  const hand = mk('speed', 'session-0f692c4e', [])
  const auto = mk('auto.0f692c4e', 'session-0f692c4e', [])
  expect(viewOf([hand, auto], [], SID).lists.map(l => l.name)).toEqual(['speed'])
})

test('手書き一覧が無ければ、現在のセッションのサイドカーの一覧を出す', async () => {
  const auto = mk('auto.0f692c4e', 'session-0f692c4e', [])
  expect(viewOf([auto], [], SID).lists.map(l => l.name)).toEqual(['auto.0f692c4e'])
})

test('新しいセッションでは、前のセッションの一覧も session の無い auto.json も出さない', async () => {
  const hand = mk('speed', 'session-0f692c4e', [task('T1', 'queued')])
  const stale = mk('auto', '', [task('T9', 'merged')])
  const old = mk('auto.0f692c4e', 'session-0f692c4e', [task('T9', 'merged')])
  const v = viewOf([hand, stale, old], [], 'ffffffff-0000-0000-0000-000000000000')
  expect(v.lists).toEqual([])
})

test('サイドカーの保存先はセッションごとに別ファイル', async () => {
  expect(autoFile('0f692c4e')).toBe('.claude/tasks/auto.0f692c4e.json')
  expect(autoFile('0f692c4e')).not.toBe(autoFile('11111111'))
})

const TASKS = [
  { list: 'a', id: 'T1', shown: 'implementing' },
  { list: 'a', id: 'T2', shown: 'queued' },
]
const ANSWER = 'T1 を main に統合し、 teammate も停止した。\nT2 はまだ着手していない。'
const change = (o: object) => JSON.stringify({ changes: [{ list: 'a', id: 'T1', status: 'merged', basis: 'T1 を main に統合し', ...o }] })

test('実在する id・既知の語・回答中の引用がそろった提案だけ通す', async () => {
  expect(parseChanges(change({}), TASKS, ANSWER)).toEqual([{ list: 'a', id: 'T1', status: 'merged', basis: 'T1 を main に統合し' }])
})

test('引用は空白や改行の違いを無視して照合する', async () => {
  expect(parseChanges(change({ basis: 'T1 を main に統合し、 teammate も\n停止した' }), TASKS, ANSWER)).toHaveLength(1)
})

// [理由, 返答]
const REJECTED: [string, string][] = [
  ['存在しない id', change({ id: 'T9' })],
  ['存在しない一覧', change({ list: 'zzz' })],
  ['未知の status の語', change({ status: 'finished' })],
  ['status が空', change({ status: '' })],
  ['basis が空', change({ basis: '' })],
  ['回答に無い basis', change({ basis: 'レビューが通ったので完了' })],
  ['対象の id を名指ししない basis', change({ basis: 'teammate も停止した' })],
  ['別タスクの文を根拠にした提案', change({ id: 'T2', status: 'done', basis: 'T1 を main に統合し' })],
  ['200字を超える basis', change({ basis: 'あ'.repeat(201) })],
  ['表示中の status と同じ', change({ status: 'implementing' })],
  ['none', 'none'],
  ['空', ''],
  ['壊れた JSON', '{"changes":['],
  ['changes が配列でない', '{"changes":{}}'],
]
for (const [why, reply] of REJECTED) {
  test(`提案を捨てる: ${why}`, async () => {
    expect(parseChanges(reply, TASKS, ANSWER)).toEqual([])
  })
}

test('コードフェンス付きの返答も読む', async () => {
  expect(parseChanges('```json\n' + change({}) + '\n```', TASKS, ANSWER)).toHaveLength(1)
})

test('1回の変更が11件を超える返答は全て捨てる', async () => {
  const many = JSON.stringify({ changes: Array.from({ length: 11 }, () => ({ list: 'a', id: 'T1', status: 'merged', basis: 'T1 を main に統合し' })) })
  expect(parseChanges(many, TASKS, ANSWER)).toEqual([])
})

test('同じタスクへの提案が重なったら最初の1件だけ通す', async () => {
  const dup = JSON.stringify({
    changes: [
      { list: 'a', id: 'T1', status: 'merged', basis: 'T1 を main に統合し' },
      { list: 'a', id: 'T1', status: 'done', basis: 'T1 を main に統合し、 teammate も停止した' },
    ],
  })
  expect(parseChanges(dup, TASKS, ANSWER).map(p => p.status)).toEqual(['merged'])
})

test('手書きの status へ戻す提案は、取り消し(status 空)として保存する', async () => {
  const p = { list: 'a', id: 'T1', status: 'implementing', basis: 'b' }
  expect(toOverlay(p, 'implementing', 'at').status).toBe('')
  expect(toOverlay({ ...p, status: 'merged' }, 'implementing', 'at')).toEqual({
    list: 'a',
    id: 'T1',
    status: 'merged',
    prev: 'implementing',
    basis: 'b',
    at: 'at',
  })
})

test('更新キーは overlay の内容で変わる', async () => {
  const base = boardKey('s', ['a\0{}'], [])
  expect(boardKey('s', ['a\0{}'], ['x\0{}'])).not.toBe(base)
  expect(boardKey('s', ['a\0{}'], [])).toBe(base)
})

test('auto の印と Pane の根拠行は、auto つきの行だけに出る', async () => {
  const plain = task('T1', 'implementing')
  const auto = { ...task('T1', 'merged'), auto: { prev: 'implementing', basis: 'T1 を統合した', at: '2026-10-05T12:00:00.000Z', list: 'a' } }
  expect(autoMark(plain)).toBe('')
  expect(autoNote(plain)).toBe('')
  expect(autoMark(auto)).toBe('auto(推定) JSON: implementing')
  expect(autoNote(auto)).toContain('a/T1 ← 「T1 を統合した」')
})

test('一覧名と id の連結が同じでも、保存先は衝突しない', async () => {
  expect(overlayPath('a__', 'b')).not.toBe(overlayPath('a', '__b'))
  expect(overlayPath('a', 'T1')).toBe(overlayPath('a', 'T1'))
})

test('長い日本語の一覧名でも、保存先のファイル名は短い', async () => {
  const name = overlayPath('あ'.repeat(60), 'T1').split('/').pop()!
  expect(name.length).toBeLessThan(40)
  expect(name).toMatch(/^[0-9a-f]+\.json$/)
})

test('間隔内に続けて終わった回答は、古い順にすべて残る', async () => {
  const p1 = queueAnswer(undefined, S('/repo/a'), 'T57 を統合した')
  const p2 = queueAnswer(p1, S('/repo/a'), 'ありがとう')
  expect(p2.answers).toEqual(['T57 を統合した', 'ありがとう'])
  expect(joinAnswers(p2.answers)).toContain('T57 を統合した')
})

test('待機中に cwd が変わったら、前の repo の回答は溜め直しで捨てる', async () => {
  const p = queueAnswer(queueAnswer(undefined, S('/repo/a'), 'A の回答'), S('/repo/b'), 'B の回答')
  expect(p).toEqual({ cwd: '/repo/b', sessionId: 'sess-1', answers: ['B の回答'] })
})

test('待機中に /clear で session ID が変わったら、前の会話の回答は溜め直しで捨てる', async () => {
  const p = queueAnswer(queueAnswer(undefined, S('/r', 'sess-1'), '旧会話の完了報告'), S('/r', 'sess-2'), '新会話の回答')
  expect(p).toEqual({ cwd: '/r', sessionId: 'sess-2', answers: ['新会話の回答'] })
})

test('auto.notes のような手書きの一覧名は、サイドカーの一覧とみなさない', async () => {
  expect(isAutoList({ name: 'auto' })).toBe(true)
  expect(isAutoList({ name: 'auto.0f692c4e' })).toBe(true)
  expect(isAutoList({ name: 'auto.notes' })).toBe(false)
  expect(isAutoList({ name: 'auto.0f692c4e.bak' })).toBe(false)
  expect(scopeLists([list('auto.notes', '')], SID).lists).toHaveLength(1)
})

test('溜める回答は8件までで、古いものから落とす', async () => {
  let p = queueAnswer(undefined, S('/r'), 'r0')
  for (let i = 1; i <= 9; i++) p = queueAnswer(p, S('/r'), `r${i}`)
  expect(p.answers).toEqual(['r2', 'r3', 'r4', 'r5', 'r6', 'r7', 'r8', 'r9'])
})

test('長い回答は結論が載りやすい末尾を残す', async () => {
  const long = 'x'.repeat(9000) + 'T1 を統合した'
  const [a] = queueAnswer(undefined, S('/r'), long).answers
  expect(a!.length).toBeLessThan(8100)
  expect(a!.endsWith('T1 を統合した')).toBe(true)
})

// [文, id, 名指しされているか]
const MENTIONS: [string, string, boolean][] = [
  ['T5 を統合した', 'T5', true],
  ['T53 を統合した', 'T5', false],
  ['T10 を統合した', 'T1', false],
  ['(T1) を統合した', 'T1', true],
  ['t1を統合した', 'T1', true],
  ['統合した: T1。', 'T1', true],
  ['XT1 を統合した', 'T1', false],
  ['#12 を修正した', '#12', true],
]
for (const [text, id, expected] of MENTIONS) {
  test(`id「${id}」は「${text}」で${expected ? '名指しされる' : '名指しされない'}`, async () => {
    expect(mentionsId(text, id)).toBe(expected)
  })
}

// [文, id, 名指しされているか]
const MENTIONS2: [string, string, boolean][] = [
  ['T5_3 を統合した', 'T5', false],
  ['T5-3 を統合した', 'T5', false],
  ['T5.2 を統合した', 'T5', false],
  ['T5を統合した', 'T5', true],
  ['T5-完了', 'T5', true],
  ['T5. 完了', 'T5', true],
]
for (const [text, id, expected] of MENTIONS2) {
  test(`id「${id}」は「${text}」で${expected ? '名指しされる' : '名指しされない'}`, async () => {
    expect(mentionsId(text, id)).toBe(expected)
  })
}

test('更新時刻は書式が混在しても最も新しい時刻を返し、時刻の無い一覧は無視する', async () => {
  const at = (updatedAt: string) => ({ ...list('x', ''), updatedAt })
  expect(latestUpdate([at('2026-10-06T19:23+0900'), at('2026-10-07T13:59:00.000Z'), at('')])).toBe('2026-10-07T13:59:00.000Z')
  expect(latestUpdate([at('')])).toBe('')
})

const COMMIT_OUT = '[main 655c2ab] feat(scripts): zmk-build\n 1 file changed, 3 insertions(+)'

test('commit が成功し `# task:T2` があれば、T2 を完了にする', async () => {
  const m = milestoneOf('git add a && git commit -m "x" # task:T2', COMMIT_OUT)
  expect(m).toEqual({ ids: ['T2'], status: 'done (commit 655c2ab)', basis: 'git commit 655c2ab' })
})

test('印にカンマ区切りで複数の id を書ける', async () => {
  expect(milestoneOf('git commit -m x # task:T2, T3', COMMIT_OUT)?.ids).toEqual(['T2', 'T3'])
})

test('印が無い commit は何も反映しない', async () => {
  expect(milestoneOf('git commit -m x', COMMIT_OUT)).toBeUndefined()
})

test('commit の出力に [branch sha] が無ければ(フックで失敗など)反映しない', async () => {
  expect(milestoneOf('git commit -m x # task:T2', 'pre-commit hook failed')).toBeUndefined()
})

test('git commit でも gh pr merge でもないコマンドは、印があっても反映しない', async () => {
  expect(milestoneOf('git push # task:T2', COMMIT_OUT)).toBeUndefined()
})

test('gh pr merge が merged を出力したら merged にする', async () => {
  expect(milestoneOf('gh pr merge 5 --squash # task:T4', '✓ Squashed and merged pull request #5')).toEqual({
    ids: ['T4'],
    status: 'merged',
    basis: 'gh pr merge',
  })
  expect(milestoneOf('gh pr merge 5 # task:T4', 'merge blocked')).toBeUndefined()
})

test('名指しされた未完了の行だけ status を差し替え、完了済みの行と他の行は触らない', async () => {
  const rows = [
    { id: 'T1', status: 'done (調査完了)' },
    { id: 'T2', status: 'implementing' },
    { id: 'T3', status: 'waiting' },
  ]
  expect(markTasks(rows, ['T1', 'T2'], 'done (commit abc1234)')).toEqual([
    { id: 'T1', status: 'done (調査完了)' },
    { id: 'T2', status: 'done (commit abc1234)' },
    { id: 'T3', status: 'waiting' },
  ])
})

test('名指しした行が全て完了済みか存在しなければ、変更なし(undefined)', async () => {
  const rows = [{ id: 'T1', status: 'merged' }]
  expect(markTasks(rows, ['T1'], 'done')).toBeUndefined()
  expect(markTasks(rows, ['T9'], 'done')).toBeUndefined()
})

test('明示の指定(overwriteDone)なら、完了済みの行も未完了へ戻せる', async () => {
  const rows = [{ id: 'T1', status: 'done (誤り)' }]
  expect(markTasks(rows, ['T1'], 'implementing')).toBeUndefined()
  expect(markTasks(rows, ['T1'], 'implementing', true)).toEqual([{ id: 'T1', status: 'implementing' }])
})

test('同じ status への差し替えは変更とみなさない', async () => {
  expect(markTasks([{ id: 'T1', status: 'implementing' }], ['T1'], 'implementing', true)).toBeUndefined()
})

test('task_set の入力は、実在の語と空でない id だけ通し、補足は括弧で後ろへ付ける', async () => {
  expect(parseSetInput({ id: ' T2 ', status: 'done' })).toEqual({ id: 'T2', status: 'done' })
  expect(parseSetInput({ id: 'T2', status: 'done', note: '調査\n完了' })).toEqual({ id: 'T2', status: 'done (調査 完了)' })
})

test('task_set の入力は、空の id・未知の語・文字列でない値を理由つきで断る', async () => {
  expect(typeof parseSetInput({ id: '', status: 'done' })).toBe('string')
  expect(parseSetInput({ id: 'T2', status: 'reviewing' })).toContain('status は')
  expect(typeof parseSetInput({ id: 2, status: 'done' })).toBe('string')
  expect(typeof parseSetInput({ id: 'T2', status: 7 })).toBe('string')
})

test('補足は100字で切る', async () => {
  const r = parseSetInput({ id: 'T2', status: 'done', note: 'あ'.repeat(300) })
  expect(typeof r === 'string' ? '' : r.status.length).toBe('done ()'.length + 100)
})

test('並列に呼んでも書き込みは呼び出し順に1本ずつ実行する', async () => {
  const log: string[] = []
  const step = (name: string, ticks: number) => async () => {
    log.push(`start ${name}`)
    for (let i = 0; i < ticks; i++) await Promise.resolve()
    log.push(`end ${name}`)
    return name
  }
  const results = await Promise.all([serialized(step('a', 30)), serialized(step('b', 1)), serialized(step('c', 1))])
  expect(results).toEqual(['a', 'b', 'c'])
  expect(log).toEqual(['start a', 'end a', 'start b', 'end b', 'start c', 'end c'])
})

test('途中の書き込みが失敗しても、後続の書き込みは実行される', async () => {
  await serialized(async () => {
    throw new Error('x')
  }).catch(() => {})
  expect(await serialized(async () => 'ok')).toBe('ok')
})

// 偽の `$`。fs はメモリ上の Map で、list は直下のファイルだけ返す(無ければ失敗)
const SID_FULL = '861bc654-aaaa-bbbb-cccc-dddddddddddd'
const fakeEnv = (files: Record<string, string>, fork?: (files: Map<string, string>) => Promise<{ isAnswered: boolean; text: string }>) => {
  const fs = new Map(Object.entries(files))
  const $: any = {
    session: { cwd: async () => '/r', id: async () => SID_FULL },
    fs: {
      list: async (dir: string) => {
        const pre = `${dir.replace(/\/$/, '')}/`
        const es = [...fs.keys()].filter(k => k.startsWith(pre) && !k.slice(pre.length).includes('/'))
        if (es.length === 0) throw new Error('ENOENT')
        return es.map(k => ({ name: k.slice(pre.length), kind: 'file', mtimeMs: 0 }))
      },
      read: async (p: string) => {
        if (!fs.has(p)) throw new Error('ENOENT')
        return fs.get(p)!
      },
      write: async (p: string, text: string) => {
        fs.set(p, text)
      },
    },
    model: { fork: async () => (fork ? fork(fs) : { isAnswered: false, text: '' }) },
  }
  return { $, fs }
}
const listJson = (...ids: string[]) =>
  JSON.stringify({ tasks: ids.map(id => ({ id, title: `t-${id}`, status: 'waiting', ws: null })) })
const ovFile = (list: string, id: string) => `/r/${overlayPath(list, id)}`

test('同じ id が複数の手書き一覧にあり list が無ければ、何も書かず一覧名を挙げて断る', async () => {
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1'), '/r/.claude/tasks/b.json': listJson('T1') })
  const r: any = await runTaskSet($, { id: 'T1', status: 'done' })
  expect(r.deny).toContain('a, b')
  expect(fs.has(ovFile('a', 'T1')) || fs.has(ovFile('b', 'T1'))).toBe(false)
})

test('list を指定すれば、その一覧の overlay だけ書き、返答に一覧名を付ける', async () => {
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1'), '/r/.claude/tasks/b.json': listJson('T1') })
  const r: any = await runTaskSet($, { id: 'T1', status: 'done', list: 'b' })
  expect(r.result).toBe('b/T1 を done にしました')
  expect(fs.has(ovFile('b', 'T1'))).toBe(true)
  expect(fs.has(ovFile('a', 'T1'))).toBe(false)
})

test('id が1つの一覧にしか無ければ、list 無しで通る', async () => {
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1'), '/r/.claude/tasks/b.json': listJson('T2') })
  const r: any = await runTaskSet($, { id: 'T2', status: 'implementing' })
  expect(r.result).toBe('b/T2 を implementing にしました')
  expect(fs.has(ovFile('b', 'T2'))).toBe(true)
})

test('一覧が1つなら返答は id だけ。無い id は、あるものを挙げて断る', async () => {
  const { $ } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1') })
  expect(((await runTaskSet($, { id: 'T1', status: 'done' })) as any).result).toBe('T1 を done にしました')
  expect(((await runTaskSet($, { id: 'T9', status: 'done' })) as any).deny).toContain('あるのは T1')
  expect(((await runTaskSet($, { id: 'T1', status: 'done', list: 'zz' })) as any).deny).toContain('一覧 zz')
})

test('fork の間に task_set が overlay を書いたら、古いサイドカーの提案は捨てて task_set の値を残す', async () => {
  const taskSet = JSON.stringify({ list: 'a', id: 'T1', status: 'implementing', prev: 'waiting', basis: 'task_set', at: '2026-10-12T00:00:00.000Z' })
  const reply = { isAnswered: true, text: JSON.stringify({ changes: [{ list: 'a', id: 'T1', status: 'done', basis: 'T1 は完了した' }] }) }
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1') }, async f => {
    f.set(ovFile('a', 'T1'), taskSet)
    return reply
  })
  const data = await loadBoard($, '/r')
  await runOverlay($, data, scopeLists(data.lists.filter(l => !isAutoList(l)), SID_FULL).lists, 'T1 は完了した', '/r')
  expect(fs.get(ovFile('a', 'T1'))).toBe(taskSet)
})

test('fork の間に何も書かれなければ、サイドカーの提案を overlay へ書く', async () => {
  const reply = { isAnswered: true, text: JSON.stringify({ changes: [{ list: 'a', id: 'T1', status: 'done', basis: 'T1 は完了した' }] }) }
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1') }, async () => reply)
  const data = await loadBoard($, '/r')
  await runOverlay($, data, scopeLists(data.lists.filter(l => !isAutoList(l)), SID_FULL).lists, 'T1 は完了した', '/r')
  expect(JSON.parse(fs.get(ovFile('a', 'T1'))!).status).toBe('done')
})

test('fork の前から overlay があり、fork の間に更新されたら、提案を捨てる', async () => {
  const before = JSON.stringify({ list: 'a', id: 'T1', status: 'running', prev: 'waiting', basis: 'old', at: '2026-10-12T00:00:00.000Z' })
  const after = JSON.stringify({ list: 'a', id: 'T1', status: 'implementing', prev: 'waiting', basis: 'task_set', at: '2026-10-12T00:00:00.000Z' })
  const reply = { isAnswered: true, text: JSON.stringify({ changes: [{ list: 'a', id: 'T1', status: 'done', basis: 'T1 は完了した' }] }) }
  const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1'), [ovFile('a', 'T1')]: before }, async f => {
    f.set(ovFile('a', 'T1'), after)
    return reply
  })
  const data = await loadBoard($, '/r')
  await runOverlay($, data, scopeLists(data.lists.filter(l => !isAutoList(l)), SID_FULL).lists, 'T1 は完了した', '/r')
  expect(fs.get(ovFile('a', 'T1'))).toBe(after)
})

test('task_set の登録は、.claude/tasks/ が無い間は呼ばず、ある repo へ移った後の poll で1回だけ呼ぶ', async () => {
  let hasDir = false
  let calls = 0
  const $: any = {
    fs: { list: async () => (hasDir ? [] : Promise.reject(new Error('ENOENT'))) },
    tool: {
      register: async () => {
        calls++
        return { tool: 'x' }
      },
    },
  }
  const state = { isRegistered: false }
  await ensureTool($, state)
  expect(calls).toBe(0)
  hasDir = true
  await ensureTool($, state)
  await ensureTool($, state)
  expect(calls).toBe(1)
})

test('task_set の登録に失敗した回は、次の poll でやり直す', async () => {
  let calls = 0
  const $: any = {
    fs: { list: async () => [] },
    tool: {
      register: async () => {
        calls++
        if (calls === 1) throw new Error('not bound')
        return { tool: 'x' }
      },
    },
  }
  const state = { isRegistered: false }
  await ensureTool($, state)
  await ensureTool($, state)
  await ensureTool($, state)
  expect(calls).toBe(2)
})

const AUTO = '/r/.claude/tasks/auto.861bc654.json'
const autoJson = () => JSON.stringify({ session: 'session-861bc654', tasks: [{ id: 'T1', title: 't', status: 'waiting', ws: null }] })

test('auto 一覧しか無いとき、存在しない list を指定した task_set は断り、ファイルを変えない', async () => {
  const { $, fs } = fakeEnv({ [AUTO]: autoJson() })
  const r: any = await runTaskSet($, { id: 'T1', status: 'done', list: 'bogus' })
  expect(r.deny).toContain('一覧 bogus')
  expect(fs.get(AUTO)).toBe(autoJson())
})

test('auto 一覧しか無いとき、list を省くか一覧名に合わせれば task_set は書ける', async () => {
  const a = fakeEnv({ [AUTO]: autoJson() })
  expect(((await runTaskSet(a.$, { id: 'T1', status: 'done' })) as any).result).toBe('T1 を done にしました')
  expect(JSON.parse(a.fs.get(AUTO)!).tasks[0].status).toBe('done')
  const b = fakeEnv({ [AUTO]: autoJson() })
  expect(((await runTaskSet(b.$, { id: 'T1', status: 'done', list: 'auto.861bc654' })) as any).result).toBe('T1 を done にしました')
})

test('commit・merge の印は、同じ id を持つ全ての手書き一覧を更新する(list の指定が無くても)', async () => {
  for (const ids of [['T1'], ['T1', 'T2']]) {
    const { $, fs } = fakeEnv({ '/r/.claude/tasks/a.json': listJson('T1'), '/r/.claude/tasks/b.json': listJson('T1', 'T2') })
    await applyStatus($, { ids, status: 'done (commit abc1234)', basis: 'git commit abc1234', isExplicit: false })
    expect(fs.has(ovFile('a', 'T1'))).toBe(true)
    expect(fs.has(ovFile('b', 'T1'))).toBe(true)
  }
})
