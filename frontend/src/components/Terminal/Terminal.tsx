import { useEffect } from 'react'
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

  // Listen for container resize
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const observer = new ResizeObserver(() => {
      handleResize()
    })

    observer.observe(container)

    return () => observer.disconnect()
  }, [handleResize, ready])

  // Auto-focus when terminal becomes active
  useEffect(() => {
    if (isActive && ready) {
      // Small delay to ensure DOM is ready
      const timer = setTimeout(() => {
        focus()
      }, 50)
      return () => clearTimeout(timer)
    }
  }, [isActive, ready, focus])

  return (
    <div
      ref={containerRef as React.RefObject<HTMLDivElement>}
      className={`w-full h-full overflow-hidden ${className}`}
      style={{ backgroundColor: '#1c2128' }}
    />
  )
}
