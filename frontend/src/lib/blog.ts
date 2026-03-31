import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const postsDirectory = path.join(process.cwd(), 'src/app/(public)/blog/content');

export interface BlogFrontmatter {
  title: string;
  date: string;
  author?: string;
  excerpt?: string;
  readingTime?: string;
  tags?: string[];
}

export interface BlogPost {
  slug: string;
  frontmatter: BlogFrontmatter;
  content: string;
}

export async function getAllBlogPosts(): Promise<BlogPost[]> {
  if (!fs.existsSync(postsDirectory)) {
    return [];
  }

  const fileNames = fs
    .readdirSync(postsDirectory)
    .filter((file) => file.endsWith('.md'));

  const posts = fileNames
    .map((fileName) => {
      const filePath = path.join(postsDirectory, fileName);
      const fileContents = fs.readFileSync(filePath, 'utf8');
      const { data, content } = matter(fileContents);

      return {
        slug: fileName.replace('.md', ''),
        frontmatter: data as BlogFrontmatter,
        content,
      };
    })
    .sort((a, b) => {
      return new Date(b.frontmatter.date).getTime() -
        new Date(a.frontmatter.date).getTime();
    });

  return posts;
}

export async function getBlogPost(slug: string): Promise<BlogPost | null> {
  const filePath = path.join(postsDirectory, `${slug}.md`);

  if (!fs.existsSync(filePath)) {
    return null;
  }

  const fileContents = fs.readFileSync(filePath, 'utf8');
  const { data, content } = matter(fileContents);

  return {
    slug,
    frontmatter: data as BlogFrontmatter,
    content,
  };
}

export function calculateReadingTime(content: string): string {
  const wordsPerMinute = 200;
  const wordCount = content.split(/\s+/).length;
  const readingTime = Math.ceil(wordCount / wordsPerMinute);
  return `${readingTime} min de leitura`;
}
