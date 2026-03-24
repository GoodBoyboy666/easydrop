import rehypeRaw from 'rehype-raw'
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize'

const mediaAttributes = [
  'autoplay',
  'controls',
  'controlsList',
  'crossorigin',
  'loop',
  'muted',
  'playsinline',
  'preload',
  'src',
] as const

const videoAttributes = ['height', 'poster', 'width'] as const

const markdownSanitizeSchema = {
  ...defaultSchema,
  tagNames: [...(defaultSchema.tagNames ?? []), 'audio', 'video'],
  attributes: {
    ...defaultSchema.attributes,
    audio: [...(defaultSchema.attributes?.audio ?? []), ...mediaAttributes],
    video: [
      ...(defaultSchema.attributes?.video ?? []),
      ...mediaAttributes,
      ...videoAttributes,
    ],
  },
  protocols: {
    ...defaultSchema.protocols,
    poster: ['http', 'https'],
    src: ['http', 'https'],
  },
}

export const markdownRehypePlugins = [
  rehypeRaw,
  [rehypeSanitize, markdownSanitizeSchema],
] as const
