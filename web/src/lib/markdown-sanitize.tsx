import rehypeRaw from 'rehype-raw'
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize'
import { BilibiliPlayer } from '#/components/markdown/bilibili-player'

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
  tagNames: [...(defaultSchema.tagNames ?? []), 'audio', 'bilibili', 'video'],
  attributes: {
    ...defaultSchema.attributes,
    audio: [...(defaultSchema.attributes?.audio ?? []), ...mediaAttributes],
    bilibili: ['bvid'],
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

export const markdownComponents = {
  bilibili: ({ bvid }: { bvid?: string }) => <BilibiliPlayer bvid={bvid} />,
}
