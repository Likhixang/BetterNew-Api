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
import { describe, test } from 'node:test'

import {
  buildTestModelOptions,
  TEST_MODEL_AUTO_VALUE,
} from '../test-model-options'

describe('buildTestModelOptions', () => {
  test('returns sorted unique model names from the models table', () => {
    const options = buildTestModelOptions([
      { model_name: 'gpt-4o' },
      { model_name: 'claude-3-5-sonnet' },
      { model_name: 'gpt-4o' },
    ])

    assert.deepEqual(options, [
      { value: 'claude-3-5-sonnet', label: 'claude-3-5-sonnet' },
      { value: 'gpt-4o', label: 'gpt-4o' },
    ])
  })

  test('merges fallback models and keeps the list populated when the table is empty', () => {
    const options = buildTestModelOptions([], ['gpt-4o', 'deepseek-chat'])

    assert.deepEqual(options, [
      { value: 'deepseek-chat', label: 'deepseek-chat' },
      { value: 'gpt-4o', label: 'gpt-4o' },
    ])
  })

  test('deduplicates table and fallback models', () => {
    const options = buildTestModelOptions(
      [{ model_name: 'gpt-4o' }],
      ['gpt-4o', 'deepseek-chat']
    )

    assert.deepEqual(options, [
      { value: 'deepseek-chat', label: 'deepseek-chat' },
      { value: 'gpt-4o', label: 'gpt-4o' },
    ])
  })

  test('includes the current value even when not present in any list', () => {
    const options = buildTestModelOptions(
      [{ model_name: 'gpt-4o' }],
      [],
      'custom-model-x'
    )

    assert.deepEqual(options, [
      { value: 'custom-model-x', label: 'custom-model-x' },
      { value: 'gpt-4o', label: 'gpt-4o' },
    ])
  })

  test('trims the current value and skips empty ones', () => {
    const options = buildTestModelOptions([{ model_name: 'gpt-4o' }], [], '   ')

    assert.deepEqual(options, [{ value: 'gpt-4o', label: 'gpt-4o' }])
  })

  test('returns an empty array when there are no models and no current value', () => {
    assert.deepEqual(buildTestModelOptions([], [], undefined), [])
    assert.deepEqual(buildTestModelOptions([], [], null), [])
  })

  test('ignores entries without a model_name', () => {
    const options = buildTestModelOptions([
      { model_name: 'gpt-4o' },
      {},
      { model_name: '' },
    ])

    assert.deepEqual(options, [{ value: 'gpt-4o', label: 'gpt-4o' }])
  })

  test('auto detect marker is a reserved value that never collides with model names', () => {
    assert.equal(TEST_MODEL_AUTO_VALUE, '__auto__')
    const options = buildTestModelOptions([], ['__auto__'])
    assert.deepEqual(options, [])
  })
})
