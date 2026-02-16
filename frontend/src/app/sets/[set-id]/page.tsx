"use client";

import { ShowFullSetsDisplay } from "@/components/shared/show-full-set";

import { usePosterSetsStore } from "@/lib/stores/global-store-poster-sets";

const SetPage = () => {
    const { setBaseInfo, posterSets, includedItems } = usePosterSetsStore();

    return (
        <ShowFullSetsDisplay
            baseSetInfo={setBaseInfo}
            posterSets={posterSets || []}
            includedItems={includedItems}
            dimNotFound={true}
        />
    );
};

export default SetPage;
