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

/**
 * Special select value that maps to an empty test_model ("auto detect").
 * The empty string itself is not used as an item value because Base UI
 * treats it as "no selection"; this marker keeps the option selectable
 * while the stored form value stays "".
 */
export const TEST_MODEL_AUTO_VALUE = '__auto__'

/**
 * Build options for the channel "Test Model" dropdown.
 *
 * Options are the models registered in "Models & Groups" (models table),
 * merged with the fallback model list (e.g. channel-level models) so the
 * dropdown stays populated even if the models table query fails.
 * The current form value is always included as a fallback so existing
 * custom entries (not in any list) stay visible and selectable.
 */
export type TestModelOption = {
  value: string
  label: string
}

export function buildTestModelOptions(
  models: Array<{ model_name?: string }>,
  fallbackModels: string[] = [],
  currentValue?: string | null
): TestModelOption[] {
  const names = new Set<string>()
  models.forEach((model) => {
    if (model.model_name && model.model_name !== TEST_MODEL_AUTO_VALUE) {
      names.add(model.model_name)
    }
  })
  fallbackModels.forEach((name) => {
    if (name && name !== TEST_MODEL_AUTO_VALUE) names.add(name)
  })
  const current = currentValue?.trim()
  if (current && current !== TEST_MODEL_AUTO_VALUE) names.add(current)
  return [...names]
    .sort((a, b) => a.localeCompare(b))
    .map((name) => ({ value: name, label: name }))
}
