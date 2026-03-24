import type { ImgHTMLAttributes } from 'react'
import type { MDEditorProps } from '@uiw/react-md-editor'
import rehypeRaw from 'rehype-raw'
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize'
import { PhotoView } from 'react-photo-view'
import remarkGfm from 'remark-gfm'
import { BilibiliPlayer } from '#/components/markdown/bilibili-player'
import { NeteasePlayer } from '#/components/markdown/netease-player'

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
    'netease',
    'video',
  ],
  attributes: {
    ...defaultSchema.attributes,
    audio: [...(defaultSchema.attributes?.audio ?? []), ...mediaAttributes],
    bilibili: ['bvid'],
    netease: ['id', 'songid'],
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
  netease: ({ id, songid }: { id?: string; songid?: string }) => (
    <NeteasePlayer id={id} songid={songid} />
  ),
}
