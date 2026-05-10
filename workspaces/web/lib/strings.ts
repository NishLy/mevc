export const generateInitials = (name: string, maxInitials = 2) => {
  const names = name.trim().split(" ")
  const initials = names
    .slice(0, maxInitials)
    .map((n) => n[0].toUpperCase())
    .join("")
  return initials || "?"
}

export const truncateString = (text: string, limit: number): string => {
  const chars = Array.from(text)

  if (chars.length <= limit) return text

  return chars.slice(0, limit).join("") + "..."
}
