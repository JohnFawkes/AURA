import React, { forwardRef } from "react";
import { cn } from "@/lib/utils";

const H1 = forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>((props, ref) => (
  <h1
    {...props}
    ref={ref}
    className={cn(
      "scroll-m-20 text-4xl font-bold lg:text-5xl mb-4",
      props.className
    )}
  >
    {props.children}
  </h1>
));
H1.displayName = "H1";
export { H1 };

const H2 = forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>((props, ref) => (
  <h2
    {...props}
    ref={ref}
    className={cn("scroll-m-20 text-3xl font-bold", props.className)}
  >
    {props.children}
  </h2>
));
H2.displayName = "H2";
export { H2 };

const H3 = forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>((props, ref) => (
  <h3
    {...props}
    ref={ref}
    className={cn("scroll-m-20 text-2xl font-semibold", props.className)}
  >
    {props.children}
  </h3>
));
H3.displayName = "H3";
export { H3 };

const H4 = forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>((props, ref) => (
  <h4
    {...props}
    ref={ref}
    className={cn(
      "scroll-m-20 text-xl font-medium tracking-wide mb-2",
      props.className
    )}
  >
    {props.children}
  </h4>
));
H4.displayName = "H4";
export { H4 };

const Lead = forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>((props, ref) => (
  <p
    {...props}
    ref={ref}
    className={cn("text-lg text-primary leading-relaxed", props.className)}
  >
    {props.children}
  </p>
));
Lead.displayName = "Lead";
export { Lead };

const P = forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>((props, ref) => (
  <p
    {...props}
    ref={ref}
    className={cn(
      "leading-7 text-base text-muted-foreground font-normal tracking-wide mb-4",
      props.className
    )}
  >
    {props.children}
  </p>
));
P.displayName = "P";
export { P };

const Small = forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>((props, ref) => (
  <p
    {...props}
    ref={ref}
    className={cn(
      "text-sm font-medium leading-none text-muted-foreground mb-1",
      props.className
    )}
  >
    {props.children}
  </p>
));
Small.displayName = "Small";
export { Small };

const InlineCode = forwardRef<
  HTMLSpanElement,
  React.HTMLAttributes<HTMLSpanElement>
>((props, ref) => (
  <code
    {...props}
    ref={ref}
    className={cn(
      "relative rounded bg-muted px-1 py-0.5 font-mono text-sm font-semibold text-foreground",
      props.className
    )}
  >
    {props.children}
  </code>
));
InlineCode.displayName = "InlineCode";
export { InlineCode };

const List = forwardRef<
  HTMLUListElement,
  React.HTMLAttributes<HTMLUListElement>
>((props, ref) => (
  <ul
    {...props}
    ref={ref}
    className={cn("my-4 ml-6 list-disc space-y-1", props.className)}
  >
    {props.children}
  </ul>
));
List.displayName = "List";
export { List };

const Quote = forwardRef<
  HTMLQuoteElement,
  React.HTMLAttributes<HTMLQuoteElement>
>((props, ref) => (
  <blockquote
    {...props}
    ref={ref}
    className={cn(
      "mt-6 border-l-4 pl-4 italic text-muted-foreground leading-relaxed",
      props.className
    )}
  >
    {props.children}
  </blockquote>
));
Quote.displayName = "Quote";
export { Quote };
