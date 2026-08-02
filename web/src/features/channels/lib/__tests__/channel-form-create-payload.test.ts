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
import { describe, test } from 'node:test'
import assert from 'node:assert/strict'

import {
  CHANNEL_FORM_DEFAULT_VALUES,
  transformFormDataToCreatePayload,
  type ChannelFormValues,
} from '../channel-form'

function createForm(
  overrides: Partial<ChannelFormValues>
): ChannelFormValues {
  return { ...CHANNEL_FORM_DEFAULT_VALUES, ...overrides }
}

describe('transformFormDataToCreatePayload mode inference', () => {
  test('single-line key with default multi_to_single mode infers single', () => {
    const payload = transformFormDataToCreatePayload(
      createForm({ key: 'sk-abc123' })
    )
    assert.equal(payload.mode, 'single')
  })

  test('multi-line key with default multi_to_single mode infers multi_to_single', () => {
    const payload = transformFormDataToCreatePayload(
      createForm({ key: 'sk-abc123\nsk-def456\nsk-ghi789' })
    )
    assert.equal(payload.mode, 'multi_to_single')
    assert.equal(payload.multi_key_mode, 'random')
  })

  test('multi-line key keeps polling strategy when selected', () => {
    const payload = transformFormDataToCreatePayload(
      createForm({
        key: 'sk-abc123\nsk-def456',
        multi_key_type: 'polling',
      })
    )
    assert.equal(payload.mode, 'multi_to_single')
    assert.equal(payload.multi_key_mode, 'polling')
  })

  test('channels forced to single mode stay single even with multi-line key', () => {
    const payload = transformFormDataToCreatePayload(
      createForm({ key: 'sk-abc123\nsk-def456', multi_key_mode: 'single' })
    )
    assert.equal(payload.mode, 'single')
  })

  test('multi-line key with blank lines still infers multi_to_single', () => {
    const payload = transformFormDataToCreatePayload(
      createForm({ key: 'sk-abc123\n\nsk-def456' })
    )
    assert.equal(payload.mode, 'multi_to_single')
  })
})
