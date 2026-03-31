import Link from 'next/link';
import { getAllBlogPosts } from '@/lib/blog';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Blog | AssinaPix',
  description: 'Dicas, guias e insights sobre assinatura, Pix recorrente e gestão de negócio.',
};

export default async function BlogPage() {
  const posts = await getAllBlogPosts();

  return (
    <div className="mx-auto max-w-4xl px-6 py-12">
      {/* Header */}
      <div className="mb-12">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">Blog AssinaPix</h1>
        <p className="text-lg text-gray-600">
          Dicas, guias e insights sobre como criar um clube de assinatura para seu negócio.
        </p>
      </div>

      {/* Posts List */}
      <div className="space-y-8">
        {posts.length === 0 ? (
          <p className="text-gray-500 text-center py-12">
            Nenhum artigo publicado ainda. Volte em breve!
          </p>
        ) : (
          posts.map((post) => (
            <article
              key={post.slug}
              className="border-b border-gray-200 pb-8 last:border-b-0"
            >
              <Link href={`/blog/${post.slug}`}>
                <h2 className="text-2xl font-bold text-gray-900 hover:text-teal-600 transition mb-2">
                  {post.frontmatter.title}
                </h2>
              </Link>
              <div className="flex items-center gap-4 text-sm text-gray-500 mb-4">
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
              </div>
              <p className="text-gray-600 mb-4 line-clamp-3">
                {post.frontmatter.excerpt || post.content.substring(0, 200) + '...'}
              </p>
              <Link
                href={`/blog/${post.slug}`}
                className="inline-flex items-center text-teal-600 hover:text-teal-700 font-medium"
              >
                Leia mais →
              </Link>
            </article>
          ))
        )}
      </div>
    </div>
  );
}
