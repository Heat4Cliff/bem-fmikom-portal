import Link from "next/link";
import { PageHeader } from "@/components/ui/primitives";
import { Breadcrumb, BreadcrumbItem } from "@/components/ui/primitives";

export function AdminPageHeader(props: { title: string; description?: string; actions?: React.ReactNode }) {
  return <PageHeader {...props} />;
}

export function AdminBreadcrumb({ items }: { items: readonly { label: string; href?: string }[] }) {
  return (
    <Breadcrumb>
      {items.map((item, index) => (
        <BreadcrumbItem key={item.href ?? item.label} current={index === items.length - 1}>
          {item.href && index < items.length - 1 ? (
            <Link href={item.href} className="hover:text-primary">
              {item.label}
            </Link>
          ) : (
            item.label
          )}
        </BreadcrumbItem>
      ))}
    </Breadcrumb>
  );
}
