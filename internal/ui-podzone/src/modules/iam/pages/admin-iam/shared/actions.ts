export type RunAction = (work: () => Promise<void>) => Promise<void>
