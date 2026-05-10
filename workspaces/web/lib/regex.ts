export const containsUrl = (text: string): boolean => {
  const urlRegex = /(https?:\/\/[^\s]+)|(www\.[^\s]+)/gi
  return urlRegex.test(text)
}

export const extractUrls = (text: string): string[] => {
  const urlRegex = /(https?:\/\/[^\s]+)|(www\.[^\s]+)/gi
  return text.match(urlRegex) || []
}

export const extractMentions = (text: string): string[] | null => {
  // Matches '@' followed by alphanumeric characters/underscores
  const mentionRegex = /@(\w+)/g
  const matches = text.match(mentionRegex)

  return matches && matches.length > 0 ? matches : null
}
