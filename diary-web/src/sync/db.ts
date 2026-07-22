import { openDB, type DBSchema, type IDBPDatabase } from 'idb';
import type { CatalogCache, LocalWorkoutSession, SyncOp } from './types';
import { SYNC_DB_NAME, SYNC_DB_VERSION, SYNC_STORE } from './types';

interface DiaryOfflineDB extends DBSchema {
  [SYNC_STORE.workoutSessions]: {
    key: string;
    value: LocalWorkoutSession;
    indexes: {
      'by-plan': string;
      'by-sync-status': string;
    };
  };
  [SYNC_STORE.syncOps]: {
    key: string;
    value: SyncOp;
    indexes: {
      'by-session-seq': [string, number];
      'by-session-status': [string, string];
      'by-status': string;
    };
  };
  [SYNC_STORE.catalog]: {
    key: string;
    value: CatalogCache;
  };
}

export type OfflineDB = IDBPDatabase<DiaryOfflineDB>;

let dbPromise: Promise<OfflineDB> | null = null;

/** Open (or reuse) the offline IndexedDB. */
export function openOfflineDB(): Promise<OfflineDB> {
  if (!dbPromise) {
    dbPromise = openDB<DiaryOfflineDB>(SYNC_DB_NAME, SYNC_DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains(SYNC_STORE.workoutSessions)) {
          const sessions = db.createObjectStore(SYNC_STORE.workoutSessions, {
            keyPath: 'id',
          });
          sessions.createIndex('by-plan', 'planId');
          sessions.createIndex('by-sync-status', 'syncStatus');
        }

        if (!db.objectStoreNames.contains(SYNC_STORE.syncOps)) {
          const ops = db.createObjectStore(SYNC_STORE.syncOps, { keyPath: 'id' });
          ops.createIndex('by-session-seq', ['sessionId', 'seq'], { unique: true });
          ops.createIndex('by-session-status', ['sessionId', 'status']);
          ops.createIndex('by-status', 'status');
        }

        if (!db.objectStoreNames.contains(SYNC_STORE.catalog)) {
          db.createObjectStore(SYNC_STORE.catalog, { keyPath: 'id' });
        }
      },
    });
  }
  return dbPromise;
}

/** Test helper: drop the shared connection so the next open is fresh. */
export async function closeOfflineDB(): Promise<void> {
  if (!dbPromise) return;
  const db = await dbPromise;
  db.close();
  dbPromise = null;
}

/** Test helper: delete the whole DB (browser / fake-indexeddb). */
export async function deleteOfflineDB(): Promise<void> {
  await closeOfflineDB();
  await new Promise<void>((resolve, reject) => {
    const req = indexedDB.deleteDatabase(SYNC_DB_NAME);
    req.onsuccess = () => resolve();
    req.onerror = () => reject(req.error ?? new Error('deleteDatabase failed'));
    req.onblocked = () => resolve();
  });
}
