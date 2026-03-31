"use client";

import { useEffect } from "react";
import { initMetaPixel } from "@/lib/analytics";

const META_PIXEL_ID = process.env.NEXT_PUBLIC_META_PIXEL_ID;

/**
 * MetaPixel component - loads and initializes Meta Pixel tracking
 * Add this component to the root layout to enable Meta Pixel tracking
 */
export default function MetaPixel() {
  useEffect(() => {
    if (META_PIXEL_ID) {
      initMetaPixel(META_PIXEL_ID);
    }
  }, []);

  return null;
}
