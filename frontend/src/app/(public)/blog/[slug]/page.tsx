import Link from 'next/link';
import { notFound } from 'next/navigation';
import { getAllBlogPosts, getBlogPost } from '@/lib/blog';
import { Metadata } from 'next';
import { marked } from 'marked';

export async function generateStaticParams() {
  const posts = await getAllBlogPosts();
  return posts.map((post) => ({
    slug: post.slug,
  }));
}

export async function generateMetadata({
  params,
}: {
  params: { slug: string };
}): Promise<Metadata> {
  const post = await getBlogPost(params.slug);
  if (!post) {
    return {
      title: 'Artigo não encontrado | ClubePay',
    };
  }
  return {
    title: `${post.frontmatter.title} | ClubePay`,
    description: post.frontmatter.excerpt || post.content.substring(0, 160),
    openGraph: {
      title: post.frontmatter.title,
      description: post.frontmatter.excerpt,
      type: 'article',
    },
  };
}

export default async function BlogPostPage({
  params,
}: {
  params: { slug: string };
}) {
  const post = await getBlogPost(params.slug);

  if (!post) {
    notFound();
  }

  const htmlContent = await marked(post.content);

  return (
    <article className="mx-auto max-w-3xl px-6 py-12">
      {/* Back Link */}
      <Link
        href="/blog"
        className="inline-flex items-center text-teal-600 hover:text-teal-700 font-medium mb-8"
      >
        ← Voltar ao blog
      </Link>

      {/* Header */}
      <header className="mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          {post.frontmatter.title}
        </h1>
        <div className="flex items-center gap-4 text-sm text-gray-500">
          <time dateTime={post.frontmatter.date}>
            {new Date(post.frontmatter.date).toLocaleDateString('pt-BR', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })}
          </time>
          {post.frontmatter.author && (
            <>
              <span>•</span>
              <span>{post.frontmatter.author}</span>
            </>
          )}
          {post.frontmatter.readingTime && (
            <>
              <span>•</span>
              <span>{post.frontmatter.readingTime}</span>
            </>
          )}
        </div>
      </header>

      {/* Content */}
      <div className="prose prose-sm max-w-none text-gray-700">
        <div
          dangerouslySetInnerHTML={{ __html: htmlContent }}
          className="space-y-4"
        />
      </div>

      {/* Footer */}
      <footer className="mt-12 pt-8 border-t border-gray-200">
        <Link
          href="/blog"
          className="inline-flex items-center text-teal-600 hover:text-teal-700 font-medium"
        >
          ← Voltar ao blog
        </Link>
      </footer>
    </article>
  );
}
