export type TEstimateType = "none" | "exponential" | "fibonacci" | "linear" | "tshirt";

const baseScales: Record<Exclude<TEstimateType, "none">, number[]> = {
  exponential: [1, 2, 4, 8, 16],
  fibonacci: [1, 2, 3, 5, 8],
  linear: [1, 2, 3, 4, 5],
  tshirt: [1, 2, 3, 5, 8],
};

const extendedScales: Record<Exclude<TEstimateType, "none">, number[]> = {
  exponential: [32, 64],
  fibonacci: [13, 21],
  linear: [6, 7],
  tshirt: [13, 21],
};

const tshirtLabels: Record<number, string> = {
  0: "None",
  1: "XS",
  2: "S",
  3: "M",
  5: "L",
  8: "XL",
  13: "XXL",
  21: "XXXL",
};

export const estimateTypeLabels: Record<TEstimateType, string> = {
  none: "Not in use",
  exponential: "Exponential (1, 2, 4, 8, 16)",
  fibonacci: "Fibonacci (1, 2, 3, 5, 8)",
  linear: "Linear (1, 2, 3, 4, 5)",
  tshirt: "T-shirt sizes (XS, S, M, L, XL)",
};

export function estimateOptions({
  type,
  allowZero,
  extended,
}: {
  type: TEstimateType;
  allowZero: boolean;
  extended: boolean;
}) {
  if (type === "none") return [];
  const values = [...baseScales[type], ...(extended ? extendedScales[type] : [])];
  return allowZero ? [0, ...values] : values;
}

export function formatEstimate(value: number, type: TEstimateType) {
  if (type === "tshirt") return tshirtLabels[value] ?? String(value);
  return String(value);
}
