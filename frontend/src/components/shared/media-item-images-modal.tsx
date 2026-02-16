"use client";

import { AssetImage } from "@/components/shared/asset-image";
import { ResponsiveGrid } from "@/components/shared/responsive-grid";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { H1, Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

interface ViewCurrentImagesModalProps {
  mediaItem: MediaItem;
  isOpen: boolean;
  onClose: () => void;
}

export function ViewCurrentImagesModal({ mediaItem, isOpen, onClose }: ViewCurrentImagesModalProps) {
  if (!mediaItem || !mediaItem.series) return null;

  const seasonRatingKeys = mediaItem.series.seasons.flatMap((season) => season.rating_key);
  const episodeRatingKeysBySeason = mediaItem.series.seasons.reduce<Record<string, string[]>>((acc, season) => {
    acc[season.rating_key] = season.episodes.map((ep) => ep.rating_key);
    return acc;
  }, {});

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className={cn("max-h-[80vh] overflow-y-auto sm:max-w-[700px]", "border border-primary")}>
        <DialogTitle className={cn("flex flex-row items-center mt-8 mb-6 px-6 w-full z-10")}>
          <div className="flex flex-col items-center mr-6">
            <AssetImage
              image={`/api/images/media/item?rating_key=${mediaItem.rating_key}&image_type=poster&cb=${Date.now()}`}
              imageType="url"
              aspect="poster"
              className="w-[140px] h-auto mb-3 rounded-lg shadow-lg"
            />
          </div>
          <div className="flex-1">
            <H1 className="text-white drop-shadow-lg text-left">{mediaItem.title}</H1>
          </div>
        </DialogTitle>
        <div className="relative w-full h-full flex flex-col items-center justify-start overflow-y-auto p-0">
          {/* Accordion for images */}
          <div className="w-full flex-1 overflow-y-auto px-4 pb-8">
            <Accordion type="multiple" className="w-full max-w-3xl mx-auto">
              <AccordionItem value="season-posters">
                <AccordionTrigger className="text-lg font-semibold text-white">Season Posters</AccordionTrigger>
                <AccordionContent>
                  <ResponsiveGrid title="">
                    {seasonRatingKeys.map((ratingKey) => (
                      <AssetImage
                        key={ratingKey}
                        image={`/api/images/media/item?rating_key=${mediaItem.rating_key}&image_rating_key=${ratingKey}&image_type=poster&cb=${Date.now()}`}
                        imageType="url"
                        aspect="poster"
                        className="w-full rounded-lg shadow"
                      />
                    ))}
                  </ResponsiveGrid>
                </AccordionContent>
              </AccordionItem>
              <AccordionItem value="titlecards">
                <AccordionTrigger className="text-lg font-semibold text-white">Episodes</AccordionTrigger>
                <AccordionContent>
                  {mediaItem?.series &&
                    Object.entries(episodeRatingKeysBySeason).map(([seasonRatingKey, episodeRatingKeys]) => (
                      <div key={seasonRatingKey} className="mb-8">
                        <Lead className="text-white mb-3 text-lg font-bold text-center">
                          {mediaItem.series?.seasons.find((season) => season.rating_key === seasonRatingKey)?.title}
                        </Lead>
                        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                          {episodeRatingKeys.map((epRatingKey) => (
                            <AssetImage
                              key={epRatingKey}
                              image={`/api/images/media/item?rating_key=${mediaItem.rating_key}&image_rating_key=${epRatingKey}&image_type=thumb&cb=${Date.now()}`}
                              imageType="url"
                              aspect="titlecard"
                              className="w-full rounded shadow"
                            />
                          ))}
                        </div>
                      </div>
                    ))}
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
