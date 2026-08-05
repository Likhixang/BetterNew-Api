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

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import {
  deleteModelMerge,
  listModelMerges,
  type ModelMerge,
} from '../lib/merge-actions'

type MergesTableProps = {
  onEdit: (merge: ModelMerge) => void
}

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

  if (isLoading) {
    return (
      <div className='flex h-40 items-center justify-center text-sm text-muted-foreground'>
        {t('Loading...')}
      </div>
    )
  }

  if (merges.length === 0) {
    return (
      <div className='flex h-40 flex-col items-center justify-center gap-2 text-sm text-muted-foreground'>
        <p>{t('No model merge rules yet.')}</p>
        <p>{t('Click "Create Model Merge" to add one.')}</p>
      </div>
    )
  }

  return (
    <div className='rounded-lg border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Target Model')}</TableHead>
            <TableHead>{t('Alias')}</TableHead>
            <TableHead>{t('Match Type')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead className='w-32 text-right'>{t('Actions')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {merges.map((merge) => (
            <TableRow key={merge.id}>
              <TableCell className='font-mono text-xs'>
                {merge.target_model}
              </TableCell>
              <TableCell className='max-w-xs truncate font-mono text-xs'>
                {merge.alias}
              </TableCell>
              <TableCell>
                {merge.match_type === 1 ? (
                  <Badge variant='outline'>{t('Regex')}</Badge>
                ) : (
                  <Badge variant='secondary'>{t('Exact')}</Badge>
                )}
              </TableCell>
              <TableCell>
                {merge.status === 1 ? (
                  <Badge className='bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
                    {t('Enabled')}
                  </Badge>
                ) : (
                  <Badge variant='secondary'>{t('Disabled')}</Badge>
                )}
              </TableCell>
              <TableCell className='text-right'>
                <div className='flex justify-end gap-2'>
                  <Button
                    variant='ghost'
                    size='sm'
                    onClick={() => props.onEdit(merge)}
                  >
                    {t('Edit')}
                  </Button>
                  <Button
                    variant='ghost'
                    size='sm'
                    className='text-destructive hover:text-destructive'
                    onClick={() => handleDelete(merge)}
                    disabled={deleteMutation.isPending}
                  >
                    {t('Delete')}
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
