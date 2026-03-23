export function normalizeMarkdownContent(value: string) {
  return value.replace(/\r\n/g, '\n').trim()
}

export function extractMarkdownText(value: string) {
  return normalizeMarkdownContent(value)
    .replace(/```([\s\S]*?)```/g, '$1')
    .replace(/`([^`]*)`/g, '$1')
    .replace(/!\[([^\]]*)\]\((.*?)\)/g, '$1')
    .replace(/\[([^\]]+)\]\((.*?)\)/g, '$1')
    .replace(/^>\s?/gm, '')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    .replace(/[*_~|]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

export function hasMarkdownContent(value: string) {
  return extractMarkdownText(value).length > 0
}
