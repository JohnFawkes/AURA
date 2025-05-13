"use client";

interface DynamicOverlayProps {
  className?: string;
}

export function DynamicOverlay({ className }: DynamicOverlayProps) {
  return (
    <div
      className={className}
      style={
        {
          background: `
            radial-gradient(ellipse at 30% 30%, var(--dynamic-left) 0%, transparent 60%),
            radial-gradient(ellipse at bottom right, var(--dynamic-bottom) 0%, transparent 60%),
            radial-gradient(ellipse at center, var(--dynamic-dark-muted) 0%, transparent 80%),
            var(--background)
          `,
          opacity: 0.5,
        } as React.CSSProperties
      }
    />
  );
}
