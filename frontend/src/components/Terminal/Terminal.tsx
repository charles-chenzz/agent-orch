import { useEffect, useRef } from 'react'
import { useTerminal } from '../../hooks/useTerminal'

interface TerminalProps {
  sessionId: string
  className?: string
  isActive?: boolean  // Whether this terminal is currently visible
}

export default function Terminal({ sessionId, className = '', isActive = true }: TerminalProps) {
  const { containerRef, handleResize, focus, ready } = useTerminal({
    sessionId,
    fontFamily: 'JetBrains Mono, Menlo, Monaco, monospace',
    fontSize: 13,
  })

  const resizeTimeoutRef = useRef<number | null>(null)
  const lastColsRef = useRef<number>(0)
  const lastRowsRef = useRef<number>(0)

  // Listen for container resize - only when active
  useEffect(() => {
    const container = containerRef.current
    if (!container || !ready) return

    const observer = new ResizeObserver(() => {
      // Only resize when terminal is visible
      if (!isActive) return

      // Debounce to avoid flickering during tab switch
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current)
      }

      resizeTimeoutRef.current = window.setTimeout(() => {
        // Get current terminal dimensions
        const terminal = containerRef.current?.querySelector('.xterm')
        if (terminal) {
          const rect = terminal.getBoundingClientRect()
          // Only resize if dimensions actually changed significantly
          if (rect.width > 0 && rect.height > 0) {
            handleResize()
          }
        }
      }, 50)
    })

    observer.observe(container)

    return () => {
      observer.disconnect()
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current)
      }
    }
  }, [handleResize, ready, isActive])

  // Handle resize when becoming active (tab switch)
  useEffect(() => {
    if (isActive && ready) {
      // Delay to ensure display: block has taken effect
      const timer = setTimeout(() => {
        handleResize()
        focus()
      }, 50)
      return () => clearTimeout(timer)
    }
  }, [isActive, ready, handleResize, focus])

  return (
    <div
      ref={containerRef as React.RefObject<HTMLDivElement>}
      className={`w-full h-full overflow-hidden ${className}`}
      style={{ backgroundColor: '#1c2128' }}
    />
  )
}
