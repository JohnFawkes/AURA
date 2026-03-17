"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

export function SelectItemsPerPage({
  setCurrentPage,
  itemsPerPage,
  setItemsPerPage,
}: {
  setCurrentPage: (page: number) => void;
  itemsPerPage: number;
  setItemsPerPage: (itemsPerPage: number) => void;
}) {
  const [customValue, setCustomValue] = useState("");

  const handleCustomSubmit = () => {
    const parsed = parseInt(customValue, 10);
    if (!isNaN(parsed) && parsed >= 1 && parsed <= 9999) {
      setItemsPerPage(parsed);
      setCurrentPage(1);
      setCustomValue("");
    }
  };

  return (
    <div className="flex flex-col gap-3">
      <Label className="text-lg font-semibold">Items per page:</Label>

      {/* Preset buttons */}
      <div className="flex flex-wrap gap-2">
        {ITEMS_PER_PAGE_OPTIONS.map((option) => (
          <Button
            key={option}
            variant="outline"
            size="sm"
            className={itemsPerPage === option ? "ring-1 ring-primary ring-offset-1" : ""}
            onClick={() => {
              setItemsPerPage(option);
              setCurrentPage(1);
              setCustomValue("");
            }}
          >
            {option}
          </Button>
        ))}
      </div>

      {/* Custom input */}
      <div className="flex items-center gap-2">
        <Input
          type="number"
          min={1}
          max={250}
          placeholder="Custom"
          value={customValue}
          className="w-28 h-8 text-sm"
          onChange={(e) => setCustomValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleCustomSubmit();
          }}
        />
        <Button variant="outline" size="sm" onClick={handleCustomSubmit} disabled={customValue === ""}>
          Apply
        </Button>
        {!ITEMS_PER_PAGE_OPTIONS.includes(itemsPerPage) && (
          <span className="text-sm text-muted-foreground">Currently: {itemsPerPage}</span>
        )}
      </div>
    </div>
  );
}
