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
import { api } from '@/lib/api'

export type ModelMerge = {
  id: number
  target_model: string
  alias: string
  match_type: 0 | 1
  status: 0 | 1
  created_time: number
  updated_time: number
}

export type ModelMergePayload = {
  id?: number
  target_model: string
  alias: string
  match_type: 0 | 1
  status: 0 | 1
}

type ApiResponse<T> = {
  success: boolean
  message: string
  data?: T
}

export async function listModelMerges(): Promise<ModelMerge[]> {
  const response = await api.get<ApiResponse<ModelMerge[]>>('/api/models/merges')
  return response.data.data ?? []
}

export async function createModelMerge(
  payload: ModelMergePayload
): Promise<ApiResponse<ModelMerge>> {
  const response = await api.post<ApiResponse<ModelMerge>>(
    '/api/models/merges',
    payload
  )
  return response.data
}

export async function updateModelMerge(
  payload: ModelMergePayload
): Promise<ApiResponse<ModelMerge>> {
  const response = await api.put<ApiResponse<ModelMerge>>(
    '/api/models/merges',
    payload
  )
  return response.data
}

export async function deleteModelMerge(id: number): Promise<ApiResponse<null>> {
  const response = await api.delete<ApiResponse<null>>(
    `/api/models/merges/${id}`
  )
  return response.data
}
