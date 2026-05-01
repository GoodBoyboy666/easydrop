import type { ImgHTMLAttributes } from 'react'
import type { MDEditorProps } from '@uiw/react-md-editor'
import rehypeRaw from 'rehype-raw'
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize'
import { PhotoView } from 'react-photo-view'
import remarkGfm from 'remark-gfm'
import { BilibiliPlayer } from '#/components/markdown/bilibili-player'
import { MetingPlayer } from '#/components/markdown/meting-player'

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
  tagNames: [
    ...(defaultSchema.tagNames ?? []),
    'audio',
    'bilibili',
    'meting',
    'video',
  ],
  attributes: {
    ...defaultSchema.attributes,
    audio: [...(defaultSchema.attributes?.audio ?? []), ...mediaAttributes],
    bilibili: ['bvid'],
    meting: ['server', 'type', 'mid'],
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

type MarkdownPreviewOptions = NonNullable<MDEditorProps['previewOptions']>
type MarkdownRehypePlugins = NonNullable<
  MarkdownPreviewOptions['rehypePlugins']
>
type MarkdownRemarkPlugins = NonNullable<
  MarkdownPreviewOptions['remarkPlugins']
>

export const markdownRehypePlugins: MarkdownRehypePlugins = [
  rehypeRaw,
  [rehypeSanitize, markdownSanitizeSchema],
]

export const markdownRemarkPlugins: MarkdownRemarkPlugins = [
  remarkGfm as MarkdownRemarkPlugins[number],
]

export const markdownComponents = {
  bilibili: ({ bvid }: { bvid?: string }) => <BilibiliPlayer bvid={bvid} />,
  img: ({
    alt,
    className,
    src,
    ...props
  }: ImgHTMLAttributes<HTMLImageElement>) => {
    if (!src) {
      return null
    }

    return (
      <PhotoView src={src}>
        <img {...props} alt={alt} className={className} src={src} />
      </PhotoView>
    )
  },
  meting: ({ server, type, mid }: { server?: string; type?: string; mid?: string }) => (
    <MetingPlayer server={server} type={type} mid={mid} />
  ),
}
