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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { DataTablePage, useDataTable } from '@/components/data-table'

import {
  deleteModelMerge,
  listModelMerges,
  type ModelMerge,
} from '../lib/merge-actions'
import { useMergesColumns } from './merges-columns'

type MergesTableProps = {
  onEdit: (merge: ModelMerge) => void
}

const DEFAULT_PAGE_SIZE = 10

export function MergesTable(props: MergesTableProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: merges = [], isLoading } = useQuery({
    queryKey: ['model-merges'],
    queryFn: listModelMerges,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteModelMerge,
    onSuccess: (response) => {
      if (response.success) {
        toast.success(t('Model merge deleted successfully'))
        queryClient.invalidateQueries({ queryKey: ['model-merges'] })
      } else {
        toast.error(response.message || t('Failed to delete model merge'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to delete model merge'))
    },
  })

  const handleDelete = (merge: ModelMerge) => {
    if (window.confirm(t('Delete this model merge rule?'))) {
      deleteMutation.mutate(merge.id)
    }
  }

  const columns = useMergesColumns({
    onEdit: props.onEdit,
    onDelete: handleDelete,
    deletePending: deleteMutation.isPending,
  })

  const { table } = useDataTable({
    data: merges,
    columns,
    initialPagination: { pageIndex: 0, pageSize: DEFAULT_PAGE_SIZE },
    withPaginationRowModel: true,
    enableSorting: false,
  })

  return (
    <DataTablePage
      table={table}
      columns={columns}
      isLoading={isLoading}
      emptyTitle={t('No model merge rules yet.')}
      emptyDescription={t('Click "Create Model Merge" to add one.')}
      toolbarProps={null}
      fixedHeight={false}
    />
  )
}
