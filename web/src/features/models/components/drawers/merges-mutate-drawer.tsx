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
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

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
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
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
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'

import {
  createModelMerge,
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

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className={sideDrawerContentClassName('sm:max-w-2xl')}>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
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
                  <Input
                    {...field}
                    placeholder='deepseek_ai/deepseek-v4-flash'
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'Exact model name or Go regex pattern (e.g. (?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash).'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

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
                    <RadioGroupItem value='0' id='merge-match-exact'>
                      {t('Exact')}
                    </RadioGroupItem>
                    <RadioGroupItem value='1' id='merge-match-regex'>
                      {t('Regex')}
                    </RadioGroupItem>
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

          <SheetFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => props.onOpenChange(false)}
            >
              {t('Close')}
            </Button>
            <Button type='submit' disabled={isSubmitting}>
              {isSubmitting ? t('Saving...') : t('Save changes')}
            </Button>
          </SheetFooter>
        </form>
      </Form>

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
