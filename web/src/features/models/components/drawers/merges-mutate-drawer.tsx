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
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import {
  sideDrawerContentClassName,
  sideDrawerFooterClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { useDebounce } from '@/hooks/use-debounce'

import {
  createModelMerge,
  previewModelMerge,
  type ModelMerge,
  updateModelMerge,
} from '../../lib/merge-actions'

type MergesMutateDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editing: ModelMerge | null
  targetModelOptions?: string[]
}

const formSchema = z.object({
  target_model: z.string().min(1),
  alias: z.string().min(1),
  match_type: z.union([z.literal(0), z.literal(1)]),
  status: z.union([z.literal(0), z.literal(1)]),
})

type FormValues = z.infer<typeof formSchema>

export function MergesMutateDrawer(props: MergesMutateDrawerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      target_model: '',
      alias: '',
      match_type: 0,
      status: 1,
    },
  })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['model-merges'] })
  }

  const createMutation = useMutation({
    mutationFn: createModelMerge,
    onSuccess: (response) => {
      if (response.success) {
        toast.success(t('Model merge created successfully'))
        invalidate()
        props.onOpenChange(false)
        form.reset()
      } else {
        toast.error(response.message || t('Failed to create model merge'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to create model merge'))
    },
  })

  const updateMutation = useMutation({
    mutationFn: updateModelMerge,
    onSuccess: (response) => {
      if (response.success) {
        toast.success(t('Model merge updated successfully'))
        invalidate()
        props.onOpenChange(false)
        form.reset()
      } else {
        toast.error(response.message || t('Failed to update model merge'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to update model merge'))
    },
  })

  // Reset form when the drawer opens with a different target (create vs edit)
  const openKey = props.editing ? `edit-${props.editing.id}` : 'create'
  const previousOpen = usePrevious(props.open)

  if (props.open && !previousOpen) {
    if (props.editing) {
      form.reset({
        target_model: props.editing.target_model,
        alias: props.editing.alias,
        match_type: props.editing.match_type,
        status: props.editing.status,
      })
    } else {
      form.reset({
        target_model: '',
        alias: '',
        match_type: 0,
        status: 1,
      })
    }
  }

  function onSubmit(values: FormValues) {
    if (props.editing) {
      updateMutation.mutate({ id: props.editing.id, ...values })
    } else {
      createMutation.mutate(values)
    }
  }

  const isSubmitting = createMutation.isPending || updateMutation.isPending

  // Live match preview: watch alias + match type, debounce, then ask the
  // backend which channels/models this draft rule would hit.
  const watchedAlias = useWatch({ control: form.control, name: 'alias' })
  const watchedMatchType = useWatch({ control: form.control, name: 'match_type' })
  const debouncedAlias = useDebounce((watchedAlias ?? '').trim(), 500)

  const { data: previewChannels = [], isFetching: previewFetching } = useQuery({
    queryKey: ['model-merge-preview', debouncedAlias, watchedMatchType],
    queryFn: () =>
      previewModelMerge({ alias: debouncedAlias, match_type: watchedMatchType }),
    enabled:
      debouncedAlias.length > 0 && (watchedMatchType === 0 || watchedMatchType === 1),
  })

  const previewError =
    debouncedAlias.length > 0 && previewChannels.length === 0
      ? t('No channels match this rule.')
      : null

  function renderPreviewContent() {
    if (debouncedAlias.length === 0) {
      return (
        <p className='mt-3 text-sm text-muted-foreground'>
          {t('Enter an alias to see matching channels and models.')}
        </p>
      )
    }
    if (previewError) {
      return <p className='mt-3 text-sm text-destructive'>{previewError}</p>
    }
    return (
      <>
        <p className='mt-3 text-xs font-medium text-muted-foreground'>
          {t('{{count}} channels matched', { count: previewChannels.length })}
        </p>
        <div className='mt-2 max-h-56 space-y-2 overflow-y-auto pr-1'>
          {previewChannels.map((channel) => (
            <div
              key={channel.id}
              className='rounded-md border bg-background p-2.5'
            >
              <div className='flex items-center gap-2'>
                <span className='truncate text-sm font-medium'>
                  {channel.name}
                </span>
                {channel.status === 1 ? (
                  <Badge className='shrink-0 bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
                    {t('Enabled')}
                  </Badge>
                ) : (
                  <Badge variant='secondary' className='shrink-0'>
                    {t('Disabled')}
                  </Badge>
                )}
              </div>
              <div className='mt-1.5 flex flex-wrap gap-1'>
                {channel.models.map((modelName) => (
                  <Badge
                    key={modelName}
                    variant='outline'
                    className='max-w-full truncate font-mono text-xs'
                  >
                    {modelName}
                  </Badge>
                ))}
              </div>
            </div>
          ))}
        </div>
      </>
    )
  }

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName('sm:max-w-2xl')}>
        <Form {...form}>
          <form
            id='model-merge-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName('gap-5')}
          >
            <SheetHeader className={sideDrawerHeaderClassName()}>
              <SheetTitle>
                {props.editing ? t('Edit Model Merge') : t('Create Model Merge')}
              </SheetTitle>
              <SheetDescription>
                {t(
                  'Merge a requested model name into a canonical target model before whitelist validation.'
                )}
              </SheetDescription>
            </SheetHeader>

          <FormField
            control={form.control}
            name='target_model'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Target Model *')}</FormLabel>
                <FormControl>
                  <Input
                    {...field}
                    list={openKey === 'create' ? 'merge-target-models' : undefined}
                    placeholder='deepseek-v4-flash'
                  />
                </FormControl>
                <FormDescription>
                  {t('The canonical model name used for whitelist, routing and billing.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='alias'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Alias *')}</FormLabel>
                <FormControl>
                  <Textarea
                    {...field}
                    className='min-h-24 resize-none font-mono text-xs'
                    placeholder={
                      'deepseek_ai/deepseek-v4-flash\n(?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash'
                    }
                    rows={3}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'One alias per line. Exact model name or Go regex pattern (e.g. (?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash).'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className='rounded-lg border bg-muted/30 p-4'>
            <div className='flex items-center justify-between'>
              <div className='space-y-1'>
                <p className='text-sm font-medium'>{t('Match Preview')}</p>
                <p className='text-xs text-muted-foreground'>
                  {t(
                    'Live preview of which channels and models this rule matches.'
                  )}
                </p>
              </div>
              {previewFetching && (
                <span className='text-xs text-muted-foreground'>
                  {t('Checking...')}
                </span>
              )}
            </div>

            {renderPreviewContent()}
          </div>

          <FormField
            control={form.control}
            name='match_type'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Match Type')}</FormLabel>
                <FormControl>
                  <RadioGroup
                    value={String(field.value)}
                    onValueChange={(value) =>
                      field.onChange(value === '1' ? 1 : 0)
                    }
                    className='flex gap-4'
                  >
                    <div className='flex items-center gap-2'>
                      <RadioGroupItem value='0' id='merge-match-exact' />
                      <Label
                        htmlFor='merge-match-exact'
                        className='cursor-pointer text-sm font-normal'
                      >
                        {t('Exact')}
                      </Label>
                    </div>
                    <div className='flex items-center gap-2'>
                      <RadioGroupItem value='1' id='merge-match-regex' />
                      <Label
                        htmlFor='merge-match-regex'
                        className='cursor-pointer text-sm font-normal'
                      >
                        {t('Regex')}
                      </Label>
                    </div>
                  </RadioGroup>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='status'
            render={({ field }) => (
              <FormItem className='flex items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <FormLabel>{t('Enabled')}</FormLabel>
                  <FormDescription>
                    {t('Disabled rules are kept but not applied.')}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value === 1}
                    onCheckedChange={(checked) =>
                      field.onChange(checked ? 1 : 0)
                    }
                  />
                </FormControl>
              </FormItem>
            )}
          />

          </form>
        </Form>

        <SheetFooter className={sideDrawerFooterClassName()}>
          <Button
            type='button'
            variant='outline'
            onClick={() => props.onOpenChange(false)}
            className='w-full sm:w-auto'
          >
            {t('Close')}
          </Button>
          <Button
            type='submit'
            form='model-merge-form'
            disabled={isSubmitting}
            className='w-full sm:w-auto'
          >
            {isSubmitting ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>

        {props.targetModelOptions && props.targetModelOptions.length > 0 && (
        <datalist id='merge-target-models'>
          {props.targetModelOptions.map((modelName) => (
            <option key={modelName} value={modelName} />
          ))}
        </datalist>
      )}
      </SheetContent>
    </Sheet>
  )
}

function usePrevious<TValue>(value: TValue): TValue | undefined {
  const ref = useRef<TValue | undefined>(undefined)
  const previous = ref.current
  ref.current = value
  return previous
}
