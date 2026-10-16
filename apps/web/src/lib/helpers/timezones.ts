const fallbackTimezones = [
  "UTC",
  "Europe/London",
  "Europe/Berlin",
  "America/New_York",
  "America/Los_Angeles",
  "Asia/Tokyo",
];

export function timezoneOptions() {
  const supported =
    typeof Intl.supportedValuesOf === "function"
      ? Intl.supportedValuesOf("timeZone")
      : fallbackTimezones;
  const zones = supported.includes("UTC") ? supported : ["UTC", ...supported];
  return zones.map((zone) => ({ value: zone, label: zone.replace(/_/g, " ") }));
}
