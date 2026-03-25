import { useEffect, useRef, useCallback, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

// Wails bindings
import {
  SendTerminalInput,
  ResizeTerminal,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { TerminalEvent } from '../types/terminal'
import { TERMINAL_EVENTS } from '../types/terminal'

interface UseTerminalOptions {
  sessionId: string
  fontFamily?: string
  fontSize?: number
  theme?: {
    background?: string
    foreground?: string
    cursor?: string
    selection?: string
  }
}

interface UseTerminalReturn {
  containerRef: React.RefObject<HTMLDivElement | null>
  terminal: React.MutableRefObject<Terminal | null>
  handleResize: () => void
  focus: () => void
  ready: boolean
}

// Default theme matching GitHub Dark Dimmed
const DEFAULT_THEME = {
  background: '#1c2128',
  foreground: '#cdd9e5',
  cursor: '#f0f6fc',
  selection: 'rgba(56, 139, 253, 0.4)',
}

export function useTerminal(options: UseTerminalOptions): UseTerminalReturn {
  const { sessionId, fontFamily = 'JetBrains Mono, Menlo, Monaco, monospace', fontSize = 13, theme = {} } = options

  const containerRef = useRef<HTMLDivElement | null>(null)
  const terminalRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const unsubscribeRef = useRef<(() => void) | null>(null)
  const [ready, setReady] = useState(false)

  // Subscribe to output events FIRST (before terminal is ready)
  // Buffer outputs until terminal is ready
  const outputBufferRef = useRef<string[]>([])

  useEffect(() => {
    if (!sessionId) return

    const handleOutput = (event: TerminalEvent) => {
      if (event.sessionId !== sessionId || !event.data) return

      if (terminalRef.current) {
        // Terminal ready, write directly
        terminalRef.current.write(event.data)
      } else {
        // Buffer output until terminal is ready
        outputBufferRef.current.push(event.data)
      }
    }

    // Subscribe immediately
    const unsubscribe = EventsOn(TERMINAL_EVENTS.OUTPUT, handleOutput)
    unsubscribeRef.current = unsubscribe

    return () => {
      if (unsubscribe) {
        unsubscribe()
      }
      unsubscribeRef.current = null
      outputBufferRef.current = []
    }
  }, [sessionId])

  // Initialize terminal
  useEffect(() => {
    if (!containerRef.current || !sessionId) return

    // Clean up existing terminal if any
    if (terminalRef.current) {
      terminalRef.current.dispose()
    }

    // Create terminal instance
    const term = new Terminal({
      fontFamily,
      fontSize,
      theme: {
        background: theme.background || DEFAULT_THEME.background,
        foreground: theme.foreground || DEFAULT_THEME.foreground,
        cursor: theme.cursor || DEFAULT_THEME.cursor,
        selectionBackground: theme.selection || DEFAULT_THEME.selection,
      },
      cursorBlink: true,
      cursorStyle: 'block',
      allowTransparency: true,
      scrollback: 10000,
    })

    // Load addons
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    term.open(containerRef.current)

    // Initial fit
    requestAnimationFrame(() => {
      fitAddon.fit()
      ResizeTerminal(sessionId, term.cols, term.rows)
    })

     // Track IME composition state using DOM events
    const isComposingRef = { current: false }
    const pendingCompositionRef = { current: '' }
    const containerEl = containerRef.current

    const handleCompositionStart = () => {
      isComposingRef.current = true
      pendingCompositionRef.current = ''
    }

    const handleCompositionUpdate = (e: CompositionEvent) => {
      // Store the current composition text (pinyin with spaces)
      pendingCompositionRef.current = e.data || ''
    }

    const handleCompositionEnd = (e: CompositionEvent) => {
      isComposingRef.current = false
      // Store final composition data
      pendingCompositionRef.current = e.data || ''
    }

    if (containerEl) {
      containerEl.addEventListener('compositionstart', handleCompositionStart)
      containerEl.addEventListener('compositionupdate', handleCompositionUpdate)
      containerEl.addEventListener('compositionend', handleCompositionEnd)
    }

    // Input handling (F2.13) - only send when not composing
    term.onData((data) => {
      // Skip input during IME composition to prevent duplicate/fragmented input
      if (isComposingRef.current) return

      // Check if this data matches the pending composition (pinyin with spaces)
      // If so, remove spaces to get raw keypresses
      if (pendingCompositionRef.current && pendingCompositionRef.current.includes(' ')) {
        // This is pinyin that wasn't committed to a candidate
        // Remove spaces to send raw keypresses (e.g., "ni hao" -> "nihao")
        const rawInput = pendingCompositionRef.current.replace(/\s+/g, '')
        pendingCompositionRef.current = ''
        if (rawInput === data.replace(/\s+/g, '')) {
          // Send the raw input without spaces
          SendTerminalInput(sessionId, rawInput)
          return
        }
      }

      SendTerminalInput(sessionId, data)
    })

    // Store refs
    terminalRef.current = term
    fitAddonRef.current = fitAddon

    // Flush buffered output to terminal
    if (outputBufferRef.current.length > 0) {
      for (const data of outputBufferRef.current) {
        term.write(data)
      }
      outputBufferRef.current = []
    }

    setReady(true)

    // Cleanup - dispose terminal only, don't call EventsOff (it affects all listeners)
    return () => {
      // Remove IME event listeners
      if (containerEl) {
        containerEl.removeEventListener('compositionstart', handleCompositionStart)
        containerEl.removeEventListener('compositionupdate', handleCompositionUpdate)
        containerEl.removeEventListener('compositionend', handleCompositionEnd)
      }
      term.dispose()
      terminalRef.current = null
      fitAddonRef.current = null
      setReady(false)
    }
  }, [sessionId, fontFamily, fontSize, theme.background, theme.foreground, theme.cursor, theme.selection])

  // Resize handler (F2.12)
  const handleResize = useCallback(() => {
    if (terminalRef.current && fitAddonRef.current && ready) {
      fitAddonRef.current.fit()
      ResizeTerminal(
        sessionId,
        terminalRef.current.cols,
        terminalRef.current.rows
      )
    }
  }, [sessionId, ready])

  // Focus
  const focus = useCallback(() => {
    terminalRef.current?.focus()
  }, [])

  return {
    containerRef,
    terminal: terminalRef,
    handleResize,
    focus,
    ready,
  }
}
