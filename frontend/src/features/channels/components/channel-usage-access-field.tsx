import { useEffect, useState } from 'react';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useSelectedProjectId } from '@/stores/projectStore';
import { apiRequest } from '@/lib/api-client';
import { extractNumberIDAsNumber } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

type Access = { mode: 'all' | 'selected'; userIDs: number[] };
type DirectoryUser = { id: number; displayName: string; isSelf: boolean };
type DirectoryPage = { users: DirectoryUser[]; nextCursor: number };

export function ChannelUsageAccessField({
  value,
  onChange,
  disabled = false,
}: {
  value?: Access | null;
  onChange: (value: Access) => void;
  disabled?: boolean;
}) {
  const { t } = useTranslation();
  const project = useSelectedProjectId();
  const currentUser = useAuthStore((state) => state.auth.user);
  const selfID = typeof currentUser?.id === 'string' ? extractNumberIDAsNumber(currentUser.id) : 0;
  const ids = value?.userIDs ?? [];
  const [search, setSearch] = useState('');
  const [query, setQuery] = useState('');
  useEffect(() => {
    const timer = setTimeout(() => setQuery(search.trim()), 250);
    return () => clearTimeout(timer);
  }, [search]);
  const load = (params: URLSearchParams) =>
    apiRequest<DirectoryPage>(`/admin/campus/access-users?${params}`, {
      requireAuth: true,
      headers: { 'X-Project-ID': project! },
    });
  const directory = useInfiniteQuery({
    queryKey: ['channelAccessUsers', project, currentUser?.id, query],
    queryFn: ({ pageParam }) => load(new URLSearchParams({ q: query, after: String(pageParam) })),
    initialPageParam: 0,
    getNextPageParam: (page) => page.nextCursor || undefined,
    enabled: !!project && value?.mode === 'selected' && !disabled,
  });
  const selected = useQuery({
    queryKey: ['channelAccessSelected', project, ids],
    queryFn: async () => {
      const users: DirectoryUser[] = [];
      // Bound URLs and follow pagination when a donor selects a large group.
      for (let i = 0; i < ids.length; i += 50) {
        const page = await load(new URLSearchParams({ ids: ids.slice(i, i + 50).join(',') }));
        users.push(...page.users);
      }
      return users;
    },
    enabled: !!project && ids.length > 0,
  });
  const names = new Map(
    [...(selected.data ?? []), ...(directory.data?.pages.flatMap((page) => page.users) ?? [])].map((user) => [user.id, user])
  );
  const label = (id: number) =>
    `${names.get(id)?.displayName ?? `#${id}`} · #${id}${id === selfID ? ` (${t('channels.usageAccess.me')})` : ''}`;
  return (
    <div className='space-y-3'>
      <Select
        value={value?.mode ?? 'all'}
        disabled={disabled}
        onValueChange={(mode) => {
          onChange(
            mode === 'self' ? { mode: 'selected', userIDs: [selfID] } : { mode: mode as Access['mode'], userIDs: mode === 'all' ? [] : ids }
          );
        }}
      >
        <SelectTrigger aria-label={t('channels.usageAccess.label')}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value='all'>{t('channels.usageAccess.all')}</SelectItem>
          <SelectItem value='selected'>{t('channels.usageAccess.selected')}</SelectItem>
          <SelectItem value='self' disabled={!selfID}>
            {t('channels.usageAccess.self')}
          </SelectItem>
        </SelectContent>
      </Select>
      <p className='text-muted-foreground text-xs'>{t('channels.usageAccess.description')}</p>
      {value?.mode === 'selected' && (
        <>
          <div className='flex flex-wrap gap-2' aria-label={t('channels.usageAccess.selected')}>
            {ids.map((id) => (
              <Button
                key={id}
                type='button'
                variant='secondary'
                size='sm'
                disabled={disabled}
                aria-label={t('channels.usageAccess.remove', { user: label(id) })}
                onClick={() => onChange({ mode: 'selected', userIDs: ids.filter((item) => item !== id) })}
              >
                {label(id)} ×
              </Button>
            ))}
          </div>
          {!disabled && (
            <>
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder={t('channels.usageAccess.search')}
                aria-label={t('channels.usageAccess.search')}
              />
              <div className='max-h-48 overflow-y-auto rounded-md border p-2'>
                {directory.data?.pages
                  .flatMap((page) => page.users)
                  .map((user) => (
                    <label key={user.id} className='hover:bg-muted flex cursor-pointer items-center gap-2 rounded px-2 py-2'>
                      <Checkbox
                        checked={ids.includes(user.id)}
                        onCheckedChange={(checked) =>
                          onChange({ mode: 'selected', userIDs: checked ? [...ids, user.id] : ids.filter((id) => id !== user.id) })
                        }
                      />
                      <span className='text-sm'>{label(user.id)}</span>
                    </label>
                  ))}
                {directory.isPending && <p className='p-2 text-sm'>{t('channels.usageAccess.loading')}</p>}
                {directory.isError && (
                  <Button type='button' variant='ghost' onClick={() => directory.refetch()}>
                    {t('channels.usageAccess.retry')}
                  </Button>
                )}
                {directory.data?.pages[0]?.users.length === 0 && <p className='p-2 text-sm'>{t('channels.usageAccess.empty')}</p>}
                {directory.hasNextPage && (
                  <Button type='button' variant='ghost' disabled={directory.isFetchingNextPage} onClick={() => directory.fetchNextPage()}>
                    {t('channels.usageAccess.more')}
                  </Button>
                )}
              </div>
            </>
          )}
          {ids.length === 0 && (
            <p className='text-destructive text-xs' role='alert'>
              {t('channels.usageAccess.required')}
            </p>
          )}
        </>
      )}
    </div>
  );
}
