const blockedPatterns = [
  /conte[uú]do sexual expl[ií]cito/i,
  /pornografia/i,
  /amea[cç]a de morte/i,
  /discurso de [oó]dio/i,
  /viol[eê]ncia expl[ií]cita/i,
];

export function hasObviouslyUnsafeContent(value: string) {
  const normalized = value.normalize('NFKC').trim();
  return blockedPatterns.some((pattern) => pattern.test(normalized));
}
