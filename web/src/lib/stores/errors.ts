import { writable } from 'svelte/store';

export interface AppError {
	id: string;
	message: string;
	type: 'error' | 'warning' | 'info';
	timestamp: number;
	details?: string;
	dismissable?: boolean;
}

function createErrorStore() {
	const { subscribe, update } = writable<AppError[]>([]);

	return {
		subscribe,
		
		/**
		 * Add an error to the store
		 */
		add: (message: string, type: AppError['type'] = 'error', details?: string) => {
			const error: AppError = {
				id: crypto.randomUUID(),
				message,
				type,
				timestamp: Date.now(),
				details,
				dismissable: true
			};
			
			update(errors => [...errors, error]);
			
			// Auto-dismiss after 10 seconds for non-critical errors
			if (type !== 'error') {
				setTimeout(() => {
					update(errors => errors.filter(e => e.id !== error.id));
				}, 10000);
			}
			
			return error.id;
		},
		
		/**
		 * Remove an error by ID
		 */
		dismiss: (id: string) => {
			update(errors => errors.filter(e => e.id !== id));
		},
		
		/**
		 * Clear all errors
		 */
		clear: () => {
			update(() => []);
		},
		
		/**
		 * Handle API response errors
		 */
		handleApiError: (response: Response, fallbackMessage = 'An error occurred') => {
			const message = `${fallbackMessage} (${response.status})`;
			return createErrorStore().add(message, 'error');
		}
	};
}

export const errors = createErrorStore();
