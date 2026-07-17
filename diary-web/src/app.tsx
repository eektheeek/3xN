import { Router } from 'preact-router';
import { WorkoutPlanList } from './pages/WorkoutPlanList';
import { WorkoutPlanNew } from './pages/WorkoutPlanNew';
import { WorkoutPlanDetail } from './pages/WorkoutPlanDetail';
import { WorkoutPlanStart } from './pages/WorkoutPlanStart';
import { ExerciseNew } from './pages/ExerciseNew';
import { ExerciseDetail } from './pages/ExerciseDetail';
import { Diary } from './pages/Diary';
import { DiarySessionDetail } from './pages/DiarySessionDetail';
import './app.css';

export function App() {
  return (
    <div class="app-shell">
      <Router>
        <WorkoutPlanList path="/" />
        <WorkoutPlanNew path="/workouts/new" />
        <WorkoutPlanStart path="/workouts/:id/start" />
        <WorkoutPlanDetail path="/workouts/:id" />
        <ExerciseNew path="/exercises/new" />
        <ExerciseDetail path="/exercises/:id" />
        <DiarySessionDetail path="/diary/:id" />
        <Diary path="/diary" />
      </Router>
    </div>
  );
}
