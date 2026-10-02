import { useState } from 'react';
import { Check, Copy, FlaskConical, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { modelAvailability, modelCandidates, orderModels } from '../data/model-availability';
import { useProbeCampusChannel, type CampusModelActivity, type CampusModelDetail, type CampusResourceChannel } from '../data/resources';

type Props = {
  models: string[];
  allModels: string[];
  details: Map<string, CampusModelDetail>;
  channels: CampusResourceChannel[];
  activity?: CampusModelActivity;
};

function ModelRow({
  model,
  rank,
  details,
  channels,
  activity,
  state,
}: {
  model: string;
  rank?: number;
  details?: CampusModelDetail;
  channels: CampusResourceChannel[];
  activity?: CampusModelActivity;
  state: ReturnType<typeof modelAvailability>;
}) {
  const { t, i18n } = useTranslation();
  const { isCopied, handleCopy } = useCopyToClipboard({ text: model });
  const [open, setOpen] = useState(false);
  const [target, setTarget] = useState('0');
  const probe = useProbeCampusChannel();
  const candidates = modelCandidates(model, channels, details).filter(({ channel }) => channel.canProbe);
  const usage = activity?.models[model];
  const format = (value: number) => value.toLocaleString(i18n.language);
  const time = (value: number) =>
    new Date(value).toLocaleString(i18n.language, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  const runTest = () => {
    const selected = candidates[Number(target)];
    if (selected?.channel.id) probe.mutate({ channelID: selected.channel.id, modelID: selected.modelID });
  };
  return (
    <article
      className='grid min-w-0 grid-cols-[minmax(0,1fr)_auto] gap-4 border-b py-5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_minmax(220px,0.85fr)_auto]'
      data-testid='campus-model-row'
      data-model={model}
    >
      <div className='min-w-0 space-y-2'>
        <button
          type='button'
          onClick={handleCopy}
          aria-label={t('resources.models.copy', { model })}
          className='group flex max-w-full items-start gap-2 rounded text-left focus-visible:outline-2'
          data-testid='campus-resource-model-copy'
          data-model-name={model}
        >
          <code className='text-sm font-semibold break-all'>{model}</code>
          {isCopied ? (
            <Check className='mt-0.5 size-4 shrink-0 text-emerald-600' />
          ) : (
            <Copy className='text-muted-foreground mt-0.5 size-4 shrink-0' />
          )}
        </button>
        <p className={state === 'tested' ? 'text-xs text-emerald-700 dark:text-emerald-400' : 'text-muted-foreground text-xs'}>
          {t(`resources.activity.state.${state}`)}
        </p>
        {details && (
          <details className='text-muted-foreground text-xs'>
            <summary className='w-fit cursor-pointer'>{t('resources.activity.capabilities')}</summary>
            <dl className='mt-2 grid grid-cols-2 gap-x-3 gap-y-1'>
              <dt>{t('resources.models.capability.source')}</dt>
              <dd>{t(`resources.models.source.${details.source}`, { defaultValue: details.source })}</dd>
              {(['vision', 'toolCall', 'reasoning'] as const).map((key) => (
                <div key={key} className='contents'>
                  <dt>{t(`resources.models.capability.${key}`)}</dt>
                  <dd>{t(`resources.models.capability.${details[key] ? 'supported' : 'unsupported'}`)}</dd>
                </div>
              ))}
              <dt>{t('resources.models.capability.context')}</dt>
              <dd>{details.contextLength > 0 ? format(details.contextLength) : t('resources.models.capability.unknown')}</dd>
              <dt>{t('resources.models.capability.output')}</dt>
              <dd>{details.maxOutputTokens === undefined ? t('resources.models.capability.unknown') : format(details.maxOutputTokens)}</dd>
            </dl>
            {details.variesByAPIKey && <p className='mt-2'>{t('resources.models.capability.variesByAPIKey')}</p>}
          </details>
        )}
      </div>
      <div className='order-3 col-span-2 space-y-2 md:order-2 md:col-span-1'>
        <div className='flex items-baseline justify-between gap-3 text-xs'>
          <span className='text-muted-foreground'>
            {t('resources.activity.successes')}{' '}
            <strong className='text-foreground font-semibold tabular-nums'>{usage ? format(usage.successCount) : '—'}</strong>
          </span>
          {rank && !!usage?.successCount && <span className='text-muted-foreground'>{t('resources.activity.rank', { rank })}</span>}
        </div>
        <div className='flex gap-1' role='img' aria-label={t('resources.activity.historyLabel', { count: usage?.successCount ?? 0 })}>
          {Array.from({ length: 24 }, (_, index) => {
            const count = usage?.buckets[index] ?? 0;
            const start = Date.parse(activity?.windowStart ?? '') + index * 3600000;
            return (
              <span
                key={index}
                className={`h-5 min-w-0 flex-1 rounded-sm ${count > 0 ? 'bg-emerald-500 dark:bg-emerald-400' : 'bg-muted'}`}
                title={
                  Number.isFinite(start)
                    ? `${time(start)} – ${time(start + 3600000)} · ${t('resources.activity.bucket', { count })}`
                    : t('resources.activity.noHistory')
                }
              />
            );
          })}
        </div>
        <div className='text-muted-foreground flex justify-between text-[10px]'>
          <span>{activity ? time(Date.parse(activity.windowStart)) : t('resources.activity.noHistory')}</span>
          <span>{activity ? time(Date.parse(activity.windowEnd)) : ''}</span>
        </div>
      </div>
      <Button
        variant='outline'
        size='sm'
        className='order-2 w-fit self-start md:order-3'
        disabled={!candidates.length}
        title={!candidates.length ? t('resources.activity.noTarget') : undefined}
        onClick={() => {
          setTarget('0');
          probe.reset();
          setOpen(true);
        }}
      >
        <FlaskConical className='size-4' />
        {t('resources.activity.test')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className='break-all'>{t('resources.activity.testTitle', { model })}</DialogTitle>
            <DialogDescription>{t('resources.activity.testScope')}</DialogDescription>
          </DialogHeader>
          <label className='space-y-2 text-sm'>
            <span>{t('resources.activity.testRoute')}</span>
            <select
              className='bg-background w-full rounded-md border p-2'
              value={target}
              disabled={probe.isPending}
              onChange={(event) => {
                setTarget(event.target.value);
                probe.reset();
              }}
            >
              {candidates.map(({ channel, modelID }, index) => (
                <option key={`${channel.id}-${modelID}`} value={index}>
                  {channel.name} · {modelID}
                </option>
              ))}
            </select>
          </label>
          <Button onClick={runTest} disabled={probe.isPending || !candidates.length}>
            {probe.isPending && <Loader2 className='size-4 animate-spin' />}
            {t(probe.isPending ? 'resources.activity.testing' : 'resources.activity.test')}
          </Button>
          <div aria-live='polite' className='space-y-2 text-sm break-words'>
            {probe.data && (
              <>
                <p>
                  {t(probe.data.success ? 'resources.activity.testPassed' : 'resources.activity.testFailed')} ·{' '}
                  {Math.round(probe.data.latency * 1000)} ms
                </p>
                {probe.data.error && <p className='text-muted-foreground'>{probe.data.error}</p>}
              </>
            )}
            {probe.error && <p>{probe.error.message}</p>}
          </div>
        </DialogContent>
      </Dialog>
    </article>
  );
}

export function ModelList({ models, allModels, details, channels, activity }: Props) {
  const { t } = useTranslation();
  const ordered = orderModels(models, activity);
  const ranks = new Map(orderModels(allModels, activity).map((model, index) => [model, index + 1]));
  const now = Date.now();
  const states = new Map(ordered.map((model) => [model, modelAvailability(model, channels, details.get(model), now, activity)]));
  const primary = ordered.filter((model) => states.get(model) !== 'failed');
  const lower = ordered.filter((model) => states.get(model) === 'failed');
  const row = (model: string) => (
    <ModelRow
      key={model}
      model={model}
      rank={ranks.get(model)}
      details={details.get(model)}
      channels={channels}
      activity={activity}
      state={states.get(model)!}
    />
  );
  return (
    <div data-testid='campus-resource-model-list'>
      <details className='text-muted-foreground mb-2 max-w-3xl text-xs leading-5'>
        <summary className='w-fit cursor-pointer'>{t('resources.activity.legend')}</summary>
        <p className='mt-2'>{t('resources.activity.explanation')}</p>
      </details>
      {primary.map(row)}
      {lower.length > 0 && (
        <details className='mt-4 rounded-lg border px-4 py-3' data-testid='campus-model-secondary'>
          <summary className='text-muted-foreground cursor-pointer text-sm'>
            {t('resources.activity.lower', { count: lower.length })}
          </summary>
          <p className='text-muted-foreground mt-2 text-xs leading-5'>{t('resources.activity.lowerExplanation')}</p>
          {lower.map(row)}
        </details>
      )}
    </div>
  );
}
