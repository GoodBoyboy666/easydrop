'use client'

import {
  FacebookIcon,
  FacebookShareButton,
  RedditIcon,
  RedditShareButton,
  TelegramIcon,
  TelegramShareButton,
  TwitterIcon,
  TwitterShareButton,
  WeiboIcon,
  WeiboShareButton,
  WhatsappIcon,
  WhatsappShareButton,
} from 'react-share'

interface PostSharePlatformsProps {
  shareTitle: string
  shareUrl: string
}

const SHARE_BUTTON_CLASS_NAME =
  'overflow-hidden rounded-full transition-transform hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'

export function PostSharePlatforms({
  shareTitle,
  shareUrl,
}: PostSharePlatformsProps) {
  return (
    <div className="grid grid-cols-3 gap-3 sm:grid-cols-6">
      <div className="flex flex-col items-center gap-1">
        <TwitterShareButton
          className={SHARE_BUTTON_CLASS_NAME}
          title={shareTitle}
          url={shareUrl}
        >
          <TwitterIcon round size={38} />
        </TwitterShareButton>
        <span className="text-xs text-muted-foreground">X</span>
      </div>

      <div className="flex flex-col items-center gap-1">
        <FacebookShareButton className={SHARE_BUTTON_CLASS_NAME} url={shareUrl}>
          <FacebookIcon round size={38} />
        </FacebookShareButton>
        <span className="text-xs text-muted-foreground">Facebook</span>
      </div>

      <div className="flex flex-col items-center gap-1">
        <TelegramShareButton
          className={SHARE_BUTTON_CLASS_NAME}
          title={shareTitle}
          url={shareUrl}
        >
          <TelegramIcon round size={38} />
        </TelegramShareButton>
        <span className="text-xs text-muted-foreground">Telegram</span>
      </div>

      <div className="flex flex-col items-center gap-1">
        <WhatsappShareButton
          className={SHARE_BUTTON_CLASS_NAME}
          separator=" "
          title={shareTitle}
          url={shareUrl}
        >
          <WhatsappIcon round size={38} />
        </WhatsappShareButton>
        <span className="text-xs text-muted-foreground">Whatsapp</span>
      </div>

      <div className="flex flex-col items-center gap-1">
        <RedditShareButton
          className={SHARE_BUTTON_CLASS_NAME}
          title={shareTitle}
          url={shareUrl}
        >
          <RedditIcon round size={38} />
        </RedditShareButton>
        <span className="text-xs text-muted-foreground">Reddit</span>
      </div>

      <div className="flex flex-col items-center gap-1">
        <WeiboShareButton
          className={SHARE_BUTTON_CLASS_NAME}
          title={shareTitle}
          url={shareUrl}
        >
          <WeiboIcon round size={38} />
        </WeiboShareButton>
        <span className="text-xs text-muted-foreground">微博</span>
      </div>
    </div>
  )
}
