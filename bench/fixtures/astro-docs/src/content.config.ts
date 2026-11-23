import { defineCollection, z } from "astro:content";
import { glob } from "astro/loaders";

export const collections = {
  docs: defineCollection({
    loader: glob({ base: "./content", pattern: "*.md" }),
    schema: z.object({
      title: z.string().min(1),
      description: z.string().min(1),
      summary: z.string().min(1),
      section: z.enum(["Guide", "Reference", "Project"]),
      source: z.string().min(1),
      order: z.number().int().min(-2147483648).max(2147483647),
    }).strict(),
  }),
};
