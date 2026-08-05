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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
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

const PAGE_SIZE = 10

type MergesTableProps = {
  onEdit: (merge: ModelMerge) => void
}

export function MergesTable(props: MergesTableProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

  const { data: merges = [], isLoading } = useQuery({
    queryKey: ['model-merges'],
    queryFn: listModelMerges,
  })

  const totalPages = Math.max(1, Math.ceil(merges.length / PAGE_SIZE))
  const currentPage = Math.min(page, totalPages)
  const pageItems = merges.slice(
    (currentPage - 1) * PAGE_SIZE,
    currentPage * PAGE_SIZE
  )

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
          {pageItems.map((merge) => (
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

      <div className='flex items-center justify-between border-t px-4 py-2'>
        <p className='text-xs text-muted-foreground'>
          {t('{{count}} rules', { count: merges.length })}
        </p>
        <Pagination className='justify-end'>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                href='#'
                onClick={(event) => {
                  event.preventDefault()
                  if (currentPage > 1) {
                    setPage(currentPage - 1)
                  }
                }}
                className={
                  currentPage <= 1 ? 'pointer-events-none opacity-50' : ''
                }
              />
            </PaginationItem>
            {Array.from({ length: totalPages }, (_, index) => index + 1).map(
              (pageNumber) => (
                <PaginationItem key={pageNumber}>
                  <PaginationLink
                    href='#'
                    onClick={(event) => {
                      event.preventDefault()
                      setPage(pageNumber)
                    }}
                    isActive={pageNumber === currentPage}
                  >
                    {pageNumber}
                  </PaginationLink>
                </PaginationItem>
              )
            )}
            <PaginationItem>
              <PaginationNext
                href='#'
                onClick={(event) => {
                  event.preventDefault()
                  if (currentPage < totalPages) {
                    setPage(currentPage + 1)
                  }
                }}
                className={
                  currentPage >= totalPages
                    ? 'pointer-events-none opacity-50'
                    : ''
                }
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    </div>
  )
}
