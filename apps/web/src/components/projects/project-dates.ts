export type TResolution = "day" | "month" | "quarter" | "half_year" | "year";

export const RESOLUTIONS: TResolution[] = ["day", "month", "quarter", "half_year", "year"];

export const resolutionLabels: Record<TResolution, string> = {
  day: "Day",
  month: "Month",
  quarter: "Quarter",
  half_year: "Half year",
  year: "Year",
};

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

function parts(date: string) {
  const [year, month, day] = date.split("-").map(Number);
  return { year, month, day };
}

export function formatProjectDate(date: string, resolution: TResolution = "day") {
  const { year, month, day } = parts(date);
  switch (resolution) {
    case "day":
      return `${MONTHS[month - 1]} ${day}, ${year}`;
    case "month":
      return `${MONTHS[month - 1]} ${year}`;
    case "quarter":
      return `Q${Math.ceil(month / 3)} ${year}`;
    case "half_year":
      return `H${month <= 6 ? 1 : 2} ${year}`;
    case "year":
      return String(year);
  }
}

function pad(value: number) {
  return String(value).padStart(2, "0");
}

export function snapToResolution(date: string, resolution: TResolution) {
  const { year, month } = parts(date);
  const endMonth = {
    day: month,
    month,
    quarter: Math.ceil(month / 3) * 3,
    half_year: month <= 6 ? 6 : 12,
    year: 12,
  }[resolution];
  if (resolution === "day") return date;
  const lastDay = new Date(Date.UTC(year, endMonth, 0)).getUTCDate();
  return `${year}-${pad(endMonth)}-${pad(lastDay)}`;
}
