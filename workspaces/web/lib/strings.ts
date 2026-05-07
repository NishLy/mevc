export const generateInitials = (name: string, maxInitials = 2) => {
  const names = name.trim().split(" ")
  const initials = names
    .slice(0, maxInitials)
    .map((n) => n[0].toUpperCase())
    .join("")
  return initials || "?"
}
