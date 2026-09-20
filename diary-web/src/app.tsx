import { useState } from 'preact/hooks';
import { Router, route } from 'preact-router';
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
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Settings } from './pages/Settings';
import { SyncDebug } from './pages/SyncDebug';
import { BottomTabBar, shouldShowTabBar } from './components/BottomTabBar';
import { DevLinksBar } from './components/DevLinksBar';
import { TunnelGateBanner } from './components/TunnelGateBanner';
import { isLoggedIn, isPublicPath } from './auth/session';
import './app.css';

export function App() {
  const [path, setPath] = useState(
    () => `${window.location.pathname}${window.location.search}`,
  );
  const [authTick, setAuthTick] = useState(0);
  const loggedIn = isLoggedIn();
  const showTabs = loggedIn && shouldShowTabBar(path);
  void authTick;

  return (
    <div class={`app-shell${showTabs ? ' app-shell--tabs' : ''}`}>
      <TunnelGateBanner />
      <div class="app-shell__content">
        <Router
          onChange={(e) => {
            const next = e.url;
            setPath(next);
            if (!isLoggedIn() && !isPublicPath(next)) {
              route('/login', true);
              return;
            }
            if (isLoggedIn() && isPublicPath(next)) {
              route('/', true);
            }
            setAuthTick((n) => n + 1);
          }}
        >
          <Login path="/login" />
          <Register path="/register" />
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
          <Settings path="/settings" />
          <SyncDebug path="/debug/sync" />
        </Router>
        {loggedIn ? <DevLinksBar /> : null}
      </div>
      {showTabs && <BottomTabBar path={path} />}
    </div>
  );
}
