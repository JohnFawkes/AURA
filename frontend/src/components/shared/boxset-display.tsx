import type { BoxsetsWithSetInfo } from "@/app/user/[username]/page";
import { setRefsToFormItems } from "@/helper/download-modal/set-to-form-item";

import DownloadModal from "@/components/shared/download-modal";
import { MediaCarousel } from "@/components/shared/media-carousel";
import { ShowFullSetsDisplay } from "@/components/shared/show-full-set";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import type { IncludedItem, SetRef } from "@/types/media-and-posters/sets";

export const RenderShowAndCollectionDisplay = ({
  set,
  includedItems,
}: {
  set: SetRef;
  includedItems?: { [tmdb_id: string]: IncludedItem };
}) => {
  // For Media Item, we find it based on the set.item_ids and the first one in the includedItems that has a title and rating key
  let mediaItem: MediaItem | undefined = undefined;
  if (includedItems) {
    for (const itemId of set.item_ids) {
      const includedItem = includedItems[itemId];
      if (includedItem && includedItem.media_item.title && includedItem.media_item.rating_key) {
        mediaItem = includedItem.media_item;
        break;
      }
    }
  }

  // Show And Collection Sets
  return (
    <MediaCarousel
      set={set}
      includedItems={includedItems}
      mediaItem={mediaItem || ({} as MediaItem)}
      dimNotFound={false}
    />
  );
};

export const RenderBoxsetDisplay = ({
  set,
  includedItems,
}: {
  set: BoxsetsWithSetInfo;
  includedItems?: { [tmdb_id: string]: IncludedItem };
}) => {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value={set.id}>
        <AccordionTrigger className="flex items-center justify-between">
          <div className="text-primary-dynamic hover:text-primary cursor-pointer text-lg font-semibold">
            {set.title}
          </div>
        </AccordionTrigger>
        <AccordionContent>
          <div className="flex justify-end">
            <DownloadModal baseSetInfo={set} formItems={setRefsToFormItems(set.sets || [], includedItems || {})} />
          </div>
          <ShowFullSetsDisplay
            baseSetInfo={set}
            posterSets={set.sets || []}
            includedItems={includedItems}
            dimNotFound={false}
          />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
};
