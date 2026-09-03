<script setup>
import { Primitive } from 'radix-vue'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

// 视觉规格与 LogShare.CN 的 AppButton 保持同步：
// 尺寸 sm=h-7 / md(默认)=h-9 / lg=h-11 / icon=h-8，变体含 soft/soft-destructive/muted 扩展
const buttonVariants = cva(
  'inline-flex select-none items-center justify-center whitespace-nowrap font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive:
          'bg-destructive text-destructive-foreground hover:bg-destructive/90',
        'soft-destructive': 'bg-destructive/10 text-destructive hover:bg-destructive/20',
        outline:
          'border border-border bg-transparent text-foreground hover:bg-accent',
        secondary:
          'bg-secondary/80 text-secondary-foreground hover:bg-secondary',
        ghost: 'bg-transparent text-muted-foreground hover:bg-accent hover:text-accent-foreground',
        soft: 'bg-primary/10 text-primary hover:bg-primary/20',
        muted: 'bg-muted text-foreground hover:bg-accent',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-9 gap-1.5 rounded-md px-4 text-sm',
        sm: 'h-7 gap-1 rounded-md px-2.5 text-xs',
        lg: 'h-11 gap-2 rounded-lg px-6 text-base',
        icon: 'h-8 w-8 rounded-md',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

defineProps({
  variant: { type: String, default: 'default' },
  size: { type: String, default: 'default' },
  as: { type: String, default: 'button' },
  class: { type: String, default: '' },
})
</script>

<template>
  <Primitive
    :as="as"
    :as-child="$attrs.asChild"
    :class="cn(buttonVariants({ variant, size }), $props.class)"
  >
    <slot />
  </Primitive>
</template>
