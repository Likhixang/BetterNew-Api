/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import assert from 'node:assert/strict'
import { after, afterEach, describe, test } from 'node:test'

import { Window } from 'happy-dom'

const { createInstance } = await import('i18next')
const { initReactI18next } = await import('react-i18next')
const { createRoot } = await import('react-dom/client')

const domWindow = new Window()
const domGlobals = [
  'window',
  'document',
  'navigator',
  'HTMLElement',
  'HTMLButtonElement',
  'HTMLInputElement',
  'SVGElement',
  'Node',
  'Element',
  'Event',
  'KeyboardEvent',
  'MouseEvent',
] as const

for (const key of domGlobals) {
  ;(globalThis as Record<string, unknown>)[key] = domWindow[key]
}

const { QueryClient, QueryClientProvider } =
  await import('@tanstack/react-query')
const { api } = await import('@/lib/api')
const { MergesTable } = await import('../merges-table')

const i18n = createInstance()
await i18n.use(initReactI18next).init({
  lng: 'en',
  resources: { en: { translation: {} } },
})

const reactTestGlobals = globalThis as typeof globalThis & {
  IS_REACT_ACT_ENVIRONMENT?: boolean
}
reactTestGlobals.IS_REACT_ACT_ENVIRONMENT = true

type ApiGet = (url: string) => Promise<{ data: unknown }>
type MockableApi = { get: ApiGet; delete?: ApiGet }
type RenderedTable = {
  host: HTMLDivElement
  queryClient: InstanceType<typeof QueryClient>
  root: ReturnType<typeof createRoot>
}

const apiClient = api as unknown as MockableApi
const originalGet = apiClient.get

function installGetMock(merges: unknown[]): void {
  apiClient.get = async (url: string) => {
    switch (url) {
      case '/api/models/merges':
        return { data: { success: true, data: merges } }
      default:
        throw new Error(`Unexpected GET ${url}`)
    }
  }
}

async function waitForCondition(
  condition: () => boolean,
  failureMessage: string
): Promise<void> {
  if (condition()) return
  const deadline = Date.now() + 3000
  while (Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 20))
    if (condition()) return
  }
  throw new Error(failureMessage)
}

async function renderTable(): Promise<RenderedTable> {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: 0 },
    },
  })
  const host = document.createElement('div')
  document.body.appendChild(host)
  const root = createRoot(host)
  root.render(
    <QueryClientProvider client={queryClient}>
      <MergesTable onEdit={() => undefined} />
    </QueryClientProvider>
  )
  await waitForCondition(
    () =>
      host.textContent?.includes('Delete') === true ||
      host.textContent?.includes('No model merge rules yet.') === true,
    'table did not render data or empty state'
  )
  return { host, queryClient, root }
}

afterEach(() => {
  apiClient.get = originalGet
})

after(async () => {
  // happy-dom window cleanup
  domWindow.close()
})

describe('MergesTable', () => {
  test('renders merge rules with alias, target and match type', async () => {
    installGetMock([
      {
        id: 1,
        target_model: 'deepseek-v4-flash',
        alias: 'deepseek_ai/deepseek-v4-flash',
        match_type: 0,
        status: 1,
        created_time: 1,
        updated_time: 1,
      },
      {
        id: 2,
        target_model: 'deepseek-v4-pro',
        alias: '(?i)(?:deepseek(?:-ai)?/)?deepseek-v4-pro',
        match_type: 1,
        status: 0,
        created_time: 1,
        updated_time: 1,
      },
    ])

    const { host } = await renderTable()

    assert.ok(host.textContent?.includes('deepseek_ai/deepseek-v4-flash'))
    assert.ok(host.textContent?.includes('deepseek-v4-flash'))
    assert.ok(
      host.textContent?.includes('(?i)(?:deepseek(?:-ai)?/)?deepseek-v4-pro')
    )
    assert.ok(host.textContent?.includes('Exact'))
    assert.ok(host.textContent?.includes('Regex'))
    assert.ok(host.textContent?.includes('Enabled'))
    assert.ok(host.textContent?.includes('Disabled'))
    assert.ok(host.textContent?.includes('Edit'))
    assert.ok(host.textContent?.includes('Delete'))
  })

  test('renders empty state when no rules exist', async () => {
    installGetMock([])

    const { host } = await renderTable()

    assert.ok(host.textContent?.includes('No model merge rules yet.'))
  })

  test('shows total rule count without pagination controls on a single page', async () => {
    const rules = Array.from({ length: 9 }, (_, index) => ({
      id: index + 1,
      target_model: `model-${index + 1}`,
      alias: `alias-${index + 1}`,
      match_type: 0,
      status: 1,
      created_time: 1,
      updated_time: 1,
    }))
    installGetMock(rules)

    const { host } = await renderTable()

    await waitForCondition(
      () => host.textContent?.includes('9 rules') === true,
      'single page should still show the total rule count'
    )
    // No page-number links when everything fits on one page
    const pageLinks = [...host.querySelectorAll('a')].filter((link) =>
      /^\d+$/.test(link.textContent?.trim() ?? '')
    )
    assert.equal(pageLinks.length, 0, 'no page number links on single page')
    assert.ok(host.textContent?.includes('alias-9'))
  })

  test('invokes onEdit with the clicked rule', async () => {
    installGetMock([
      {
        id: 7,
        target_model: 'gpt-5.5',
        alias: 'cx/gpt-5.5',
        match_type: 0,
        status: 1,
        created_time: 1,
        updated_time: 1,
      },
    ])

    let editedId: number | null = null
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: 0 } },
    })
    const host = document.createElement('div')
    document.body.appendChild(host)
    const root = createRoot(host)
    root.render(
      <QueryClientProvider client={queryClient}>
        <MergesTable onEdit={(merge) => (editedId = merge.id)} />
      </QueryClientProvider>
    )
    await waitForCondition(
      () => host.textContent?.includes('cx/gpt-5.5') === true,
      'rule row did not render'
    )

    const buttons = [...host.querySelectorAll('button')]
    const editButton = buttons.find((button) => button.textContent === 'Edit')
    assert.ok(editButton, 'Edit button should exist')
    editButton.click()
    assert.equal(editedId, 7)
  })

  test('paginates 13 rules into 10 + 3 with page navigation', async () => {
    const rules = Array.from({ length: 13 }, (_, index) => ({
      id: index + 1,
      target_model: `model-${index + 1}`,
      alias: `alias-${index + 1}`,
      match_type: 0,
      status: 1,
      created_time: 1,
      updated_time: 1,
    }))
    installGetMock(rules)

    const { host } = await renderTable()

    // First page shows 10 rows
    await waitForCondition(
      () => host.textContent?.includes('alias-10') === true,
      'first page should show up to 10 rules'
    )
    assert.ok(host.textContent?.includes('alias-1'))
    assert.ok(host.textContent?.includes('alias-10'))
    assert.ok(!host.textContent?.includes('alias-11'))

    // Navigate to page 2
    const pageButtons = [...host.querySelectorAll('a')].filter((link) =>
      link.textContent?.trim()
    )
    const pageTwo = pageButtons.find((link) => link.textContent?.trim() === '2')
    assert.ok(pageTwo, 'page 2 link should exist')
    pageTwo.click()

    await waitForCondition(
      () => host.textContent?.includes('alias-11') === true,
      'second page should show remaining rules'
    )
    assert.ok(host.textContent?.includes('alias-13'))
    // alias-1 followed by a non-digit must not appear (alias-13 contains the
    // substring alias-1, so use a boundary check)
    assert.ok(
      !/alias-1(?![0-9])/.test(host.textContent ?? ''),
      'first page rule alias-1 should not appear on page 2'
    )
  })
})
