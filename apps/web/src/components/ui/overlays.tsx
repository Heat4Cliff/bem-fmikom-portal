"use client";

import type * as React from "react";
import * as AlertDialogPrimitive from "@radix-ui/react-alert-dialog";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import * as ToastPrimitive from "@radix-ui/react-toast";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import { Check, ChevronRight, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils/cn";

export const Dialog = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;

export function DialogContent({
  className,
  children,
  ...props
}: React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-slate-950/45 backdrop-blur-[2px] data-[state=closed]:animate-out data-[state=open]:animate-in" />
      <DialogPrimitive.Content
        className={cn(
          "fixed left-1/2 top-1/2 z-50 max-h-[min(90vh,48rem)] w-[min(92vw,38rem)] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-xl border border-border bg-surface p-6 shadow-float focus:outline-none",
          className,
        )}
        {...props}
      >
        {children}
        <DialogPrimitive.Close
          aria-label="Tutup dialog"
          className="absolute right-4 top-4 grid size-9 place-items-center rounded-md text-muted hover:bg-surface-muted hover:text-primary focus-visible:outline-2 focus-visible:outline-secondary"
        >
          <X aria-hidden="true" className="size-4" />
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  );
}

export const DialogTitle = DialogPrimitive.Title;
export const DialogDescription = DialogPrimitive.Description;
export const DialogClose = DialogPrimitive.Close;

export function Drawer({
  open,
  onOpenChange,
  title,
  side = "left",
  children,
  trigger,
  className,
}: {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: string;
  side?: "left" | "right";
  children: React.ReactNode;
  trigger?: React.ReactNode;
  className?: string;
}) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      {trigger ? <DialogPrimitive.Trigger asChild>{trigger}</DialogPrimitive.Trigger> : null}
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-slate-950/45" />
        <DialogPrimitive.Content
          className={cn(
            "fixed inset-y-0 z-50 flex w-[min(88vw,22rem)] flex-col bg-surface shadow-float focus:outline-none",
            side === "left" ? "left-0 border-r border-border" : "right-0 border-l border-border",
            className,
          )}
        >
          <DialogPrimitive.Title className="sr-only">{title}</DialogPrimitive.Title>
          <div className="flex items-center justify-between border-b border-border px-5 py-4">
            <span className="font-heading font-bold text-primary">{title}</span>
            <DialogPrimitive.Close
              aria-label="Tutup navigasi"
              className="grid size-10 place-items-center rounded-md text-muted hover:bg-surface-muted"
            >
              <X aria-hidden="true" className="size-5" />
            </DialogPrimitive.Close>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto">{children}</div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

export function ConfirmationDialog({
  title,
  description,
  confirmLabel = "Konfirmasi",
  cancelLabel = "Batal",
  onConfirm,
  children,
  destructive = false,
}: {
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  children: React.ReactNode;
  destructive?: boolean;
}) {
  return (
    <AlertDialogPrimitive.Root>
      <AlertDialogPrimitive.Trigger asChild>{children}</AlertDialogPrimitive.Trigger>
      <AlertDialogPrimitive.Portal>
        <AlertDialogPrimitive.Overlay className="fixed inset-0 z-50 bg-slate-950/45" />
        <AlertDialogPrimitive.Content className="fixed left-1/2 top-1/2 z-50 w-[min(92vw,28rem)] -translate-x-1/2 -translate-y-1/2 rounded-xl border border-border bg-surface p-6 shadow-float focus:outline-none">
          <AlertDialogPrimitive.Title className="font-heading text-lg font-bold text-primary">
            {title}
          </AlertDialogPrimitive.Title>
          <AlertDialogPrimitive.Description className="mt-2 text-sm leading-6 text-muted">
            {description}
          </AlertDialogPrimitive.Description>
          <div className="mt-6 flex justify-end gap-2">
            <AlertDialogPrimitive.Cancel asChild>
              <Button variant="outline">{cancelLabel}</Button>
            </AlertDialogPrimitive.Cancel>
            <AlertDialogPrimitive.Action asChild>
              <Button variant={destructive ? "danger" : "primary"} onClick={onConfirm}>
                {confirmLabel}
              </Button>
            </AlertDialogPrimitive.Action>
          </div>
        </AlertDialogPrimitive.Content>
      </AlertDialogPrimitive.Portal>
    </AlertDialogPrimitive.Root>
  );
}

export function DropdownMenu({
  trigger,
  items,
}: {
  trigger: React.ReactNode;
  items: readonly { label: string; onSelect: () => void; disabled?: boolean; destructive?: boolean }[];
}) {
  return (
    <DropdownMenuPrimitive.Root>
      <DropdownMenuPrimitive.Trigger asChild>{trigger}</DropdownMenuPrimitive.Trigger>
      <DropdownMenuPrimitive.Portal>
        <DropdownMenuPrimitive.Content
          align="end"
          sideOffset={6}
          className="z-50 min-w-48 rounded-md border border-border bg-surface p-1.5 shadow-float"
        >
          {items.map((item) => (
            <DropdownMenuPrimitive.Item
              key={item.label}
              disabled={item.disabled}
              onSelect={item.onSelect}
              className={cn(
                "flex min-h-10 cursor-pointer items-center justify-between rounded px-3 text-sm text-body outline-none focus:bg-surface-muted data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
                item.destructive && "text-danger focus:bg-danger/5",
              )}
            >
              {item.label}
              <ChevronRight aria-hidden="true" className="size-3.5 opacity-50" />
            </DropdownMenuPrimitive.Item>
          ))}
        </DropdownMenuPrimitive.Content>
      </DropdownMenuPrimitive.Portal>
    </DropdownMenuPrimitive.Root>
  );
}

export function Tooltip({ content, children }: { content: string; children: React.ReactNode }) {
  return (
    <TooltipPrimitive.Provider delayDuration={300}>
      <TooltipPrimitive.Root>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            sideOffset={6}
            className="z-50 rounded-md bg-primary px-2.5 py-1.5 text-xs font-medium text-primary-foreground shadow-float"
          >
            {content}
            <TooltipPrimitive.Arrow className="fill-primary" />
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  );
}

export function Toast({
  open,
  onOpenChange,
  title,
  description,
  variant = "success",
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  variant?: "success" | "danger" | "info";
}) {
  return (
    <ToastPrimitive.Provider swipeDirection="right">
      <ToastPrimitive.Root
        open={open}
        onOpenChange={onOpenChange}
        className={cn(
          "grid w-full max-w-sm grid-cols-[1fr_auto] gap-x-3 rounded-lg border bg-surface p-4 shadow-float data-[state=open]:animate-in",
          variant === "success" && "border-success/40",
          variant === "danger" && "border-danger/40",
          variant === "info" && "border-info/40",
        )}
      >
        <div>
          <ToastPrimitive.Title className="text-sm font-bold text-primary">{title}</ToastPrimitive.Title>
          {description ? (
            <ToastPrimitive.Description className="mt-1 text-sm leading-5 text-muted">
              {description}
            </ToastPrimitive.Description>
          ) : null}
        </div>
        <ToastPrimitive.Close
          aria-label="Tutup notifikasi"
          className="grid size-8 place-items-center rounded text-muted hover:bg-surface-muted"
        >
          <Check aria-hidden="true" className="size-4" />
        </ToastPrimitive.Close>
      </ToastPrimitive.Root>
      <ToastPrimitive.Viewport className="fixed bottom-4 right-4 z-[100] flex w-[min(92vw,24rem)] flex-col gap-2 outline-none" />
    </ToastPrimitive.Provider>
  );
}
