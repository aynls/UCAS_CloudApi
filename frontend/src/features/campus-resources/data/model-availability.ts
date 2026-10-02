import type { CampusModelActivity, CampusModelDetail, CampusResourceChannel } from './resources';

export const TEST_FRESHNESS_MS = 60 * 60 * 1000;

export function modelCandidates(model: string, channels: CampusResourceChannel[], details?: CampusModelDetail) {
  const targets = details?.testTargets.length
    ? details.testTargets
    : channels
        .filter((channel) => channel.models.includes(model) && channel.id)
        .map((channel) => ({ channelID: channel.id!, modelID: model }));
  return targets.flatMap((target) => {
    const channel = channels.find((entry) => entry.id === target.channelID && entry.status === 'enabled');
    return channel ? [{ channel, modelID: target.modelID }] : [];
  });
}

export function modelAvailability(
  model: string,
  channels: CampusResourceChannel[],
  details: CampusModelDetail | undefined,
  now: number,
  activity?: CampusModelActivity
) {
  const candidates = modelCandidates(model, channels, details);
  // Older servers lack route metadata. Missing matches are unknown, never proof of failure.
  if (!candidates.length) return 'unknown' as const;
  let allFailed = !details?.testTargets.length || details.testTargets.length === candidates.length;
  let earliestFailure = Infinity;
  for (const { channel, modelID } of candidates) {
    const routes = channel.health?.routes.filter((route) => route.requestModels?.includes(modelID) || route.model === modelID) ?? [];
    if (!routes.length) allFailed = false;
    for (const route of routes) {
      const at = Date.parse(route.lastTestAt ?? '');
      const fresh = route.known && !route.testInFlight && at <= now && now - at < TEST_FRESHNESS_MS;
      if (fresh && route.available) return 'tested' as const;
      if (!fresh) allFailed = false;
      if (fresh && !route.available) earliestFailure = Math.min(earliestFailure, at);
    }
  }
  // Coarse history may contain a success newer than the negative test. Keep it discoverable.
  const usage = activity?.models[model];
  if (allFailed && usage && activity) {
    const lastSuccessBucket = usage.buckets.findLastIndex((count) => count > 0);
    const bucketEnd = Date.parse(activity.windowStart) + (lastSuccessBucket + 1) * activity.bucketHours * TEST_FRESHNESS_MS;
    if (lastSuccessBucket >= 0 && bucketEnd >= earliestFailure) allFailed = false;
  }
  return allFailed ? ('failed' as const) : ('unknown' as const);
}

export function orderModels(models: string[], activity?: CampusModelActivity) {
  return [...new Set(models)].sort(
    (a, b) => (activity?.models[b]?.successCount ?? 0) - (activity?.models[a]?.successCount ?? 0) || a.localeCompare(b)
  );
}
