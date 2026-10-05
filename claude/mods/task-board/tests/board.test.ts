import { test, expect } from 'claude-code/testing'

import { kindOf, note, scopeLists } from '../hooks/register'

const t = (status: string) => ({ id: 'T1', title: 'x', status, ws: null })

// [status, 分類, 補足]
const CASES: [string, string, string][] = [
  ['done', 'done', ''],
  ['done: x', 'done', ' x'],
  ['done：x', 'done', ' x'],
  ['done: 該当0件', 'done', ' 該当0件'],
  ['merged', 'done', ''],
  ['implementing', 'active', ''],
  ['fixing', 'active', ''],
  ['running', 'active', ''],
  ['waiting (T1後)', 'open', ' (T1後)'],
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

test('現在のセッションに一致する一覧だけに絞る', async () => {
  const r = scopeLists([A, B, C], '0f692c4e-aaaa-bbbb-cccc-dddddddddddd')
  expect(r.scoped).toBe(true)
  expect(r.lists).toEqual([A])
})

test('一致する一覧が無ければ全件を返し、絞り込み中にならない', async () => {
  const r = scopeLists([B, C], '0f692c4e-aaaa-bbbb-cccc-dddddddddddd')
  expect(r.scoped).toBe(false)
  expect(r.lists).toEqual([B, C])
})

test('セッションIDが空なら全件を返す', async () => {
  expect(scopeLists([A, C], '').scoped).toBe(false)
})
