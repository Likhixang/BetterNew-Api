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
import type { ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

import type { ModelMerge } from '../lib/merge-actions'

type MergesColumnOptions = {
  onEdit: (merge: ModelMerge) => void
  onDelete: (merge: ModelMerge) => void
  deletePending: boolean
}

export function useMergesColumns(
  options: MergesColumnOptions
): ColumnDef<ModelMerge>[] {
  const { t } = useTranslation()
  return [
    {
      accessorKey: 'target_model',
      header: () => t('Target Model'),
      meta: { label: t('Target Model') },
      cell: ({ row }) => (
        <span className='font-mono text-xs'>{row.original.target_model}</span>
      ),
      size: 200,
    },
    {
      accessorKey: 'alias',
      header: () => t('Alias'),
      meta: { label: t('Alias') },
      cell: ({ row }) => (
        <span className='block max-w-xs truncate font-mono text-xs'>
          {row.original.alias}
        </span>
      ),
      size: 320,
    },
    {
      accessorKey: 'match_type',
      header: () => t('Match Type'),
      meta: { label: t('Match Type') },
      cell: ({ row }) =>
        row.original.match_type === 1 ? (
          <Badge variant='outline'>{t('Regex')}</Badge>
        ) : (
          <Badge variant='secondary'>{t('Exact')}</Badge>
        ),
      enableSorting: false,
      size: 110,
    },
    {
      accessorKey: 'status',
      header: () => t('Status'),
      meta: { label: t('Status') },
      cell: ({ row }) =>
        row.original.status === 1 ? (
          <Badge className='bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
            {t('Enabled')}
          </Badge>
        ) : (
          <Badge variant='secondary'>{t('Disabled')}</Badge>
        ),
      enableSorting: false,
      size: 100,
    },
    {
      id: 'actions',
      header: () => t('Actions'),
      cell: ({ row }) => (
        <div className='flex justify-end gap-2'>
          <Button
            variant='ghost'
            size='sm'
            onClick={() => options.onEdit(row.original)}
          >
            {t('Edit')}
          </Button>
          <Button
            variant='ghost'
            size='sm'
            className='text-destructive hover:text-destructive'
            onClick={() => options.onDelete(row.original)}
            disabled={options.deletePending}
          >
            {t('Delete')}
          </Button>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
      meta: { pinned: 'right' as const },
    },
  ]
}
