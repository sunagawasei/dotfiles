import { test, expect } from 'claude-code/testing'

import { kindOf, note } from '../hooks/register'

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
