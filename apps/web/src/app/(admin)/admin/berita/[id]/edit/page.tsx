import { notFound } from "next/navigation";
import { ArticleEditor } from "@/features/content/article-editor";

export default async function EditArticlePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const articleID = Number.parseInt(id, 10);
  if (!Number.isSafeInteger(articleID) || articleID < 1) notFound();
  return <ArticleEditor id={articleID} />;
}
