import TopPanel from './components/Layout/TopPanel'
import TerminalPane from './components/Layout/TerminalPane'
import DiffPanel from './components/Layout/DiffPanel'
import StatusBar from './components/Layout/StatusBar'

function App() {
  return (
    <div className="flex flex-col h-screen bg-base text-text overflow-hidden">
      <TopPanel />
      <TerminalPane />
      <DiffPanel />
      <StatusBar />
    </div>
  )
}

export default App
