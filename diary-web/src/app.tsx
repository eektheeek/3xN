import { useState } from 'preact/hooks';
import { Router } from 'preact-router';
import { WorkoutPlanList } from './pages/WorkoutPlanList';
import { WorkoutPlanNew } from './pages/WorkoutPlanNew';
import { WorkoutPlanDetail } from './pages/WorkoutPlanDetail';
import { WorkoutPlanStart } from './pages/WorkoutPlanStart';
import { ExerciseNew } from './pages/ExerciseNew';
import { ExerciseDetail } from './pages/ExerciseDetail';
import { Diary } from './pages/Diary';
import { DiarySessionDetail } from './pages/DiarySessionDetail';
import { ProtocolList } from './pages/ProtocolList';
import { ProtocolNew } from './pages/ProtocolNew';
import { ProtocolDetail } from './pages/ProtocolDetail';
import { CycleList } from './pages/CycleList';
import { CycleDetail } from './pages/CycleDetail';
import { Stats } from './pages/Stats';
import { More } from './pages/More';
import { SyncDebug } from './pages/SyncDebug';
import { BottomTabBar, shouldShowTabBar } from './components/BottomTabBar';
import { DevLinksBar } from './components/DevLinksBar';
import { TunnelGateBanner } from './components/TunnelGateBanner';
import './app.css';

export function App() {
  const [path, setPath] = useState(
    () => `${window.location.pathname}${window.location.search}`,
  );
  const showTabs = shouldShowTabBar(path);

  return (
    <div class={`app-shell${showTabs ? ' app-shell--tabs' : ''}`}>
      <TunnelGateBanner />
      <div class="app-shell__content">
        <Router onChange={(e) => setPath(e.url)}>
          <WorkoutPlanList path="/" />
          <WorkoutPlanNew path="/workouts/new" />
          <WorkoutPlanStart path="/workouts/:id/start" />
          <WorkoutPlanDetail path="/workouts/:id" />
          <ExerciseNew path="/exercises/new" />
          <ExerciseDetail path="/exercises/:id" />
          <ProtocolNew path="/protocols/new" />
          <ProtocolDetail path="/protocols/:id" />
          <ProtocolList path="/protocols" />
          <CycleList path="/cycles" />
          <CycleDetail path="/cycles/:id" />
          <DiarySessionDetail path="/diary/:id" />
          <Diary path="/diary" />
          <Stats path="/stats" />
          <More path="/more" />
          <SyncDebug path="/debug/sync" />
        </Router>
        <DevLinksBar />
      </div>
      {showTabs && <BottomTabBar path={path} />}
    </div>
  );
}
