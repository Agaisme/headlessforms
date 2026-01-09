/**
 * Submissions Repository - API abstraction for submission operations
 */

import { api, type ApiResponse } from './client';

export interface Submission {
	id: string;
	form_id: string;
	data: Record<string, unknown>;
	meta?: Record<string, unknown>;
	is_read: boolean;
	is_spam: boolean;
	created_at: string;
}

export const submissionsRepository = {
	/**
	 * List submissions for a form
	 */
	list: async (formId: string, page = 1, limit = 50) => {
		return api.get<{ submissions: Submission[]; pagination: unknown }>(
			`/forms/${formId}/submissions?page=${page}&limit=${limit}`
		);
	},

	/**
	 * Get a single submission
	 */
	get: async (submissionId: string) => {
		return api.get<Submission>(`/submissions/${submissionId}`);
	},

	/**
	 * Mark submission as read
	 */
	markAsRead: async (submissionId: string) => {
		return api.put<void>(`/submissions/${submissionId}/read`);
	},

	/**
	 * Mark submission as unread
	 */
	markAsUnread: async (submissionId: string) => {
		return api.put<void>(`/submissions/${submissionId}/unread`);
	},

	/**
	 * Delete a submission
	 */
	delete: async (submissionId: string) => {
		return api.delete<void>(`/submissions/${submissionId}`);
	},

	/**
	 * Export submissions as CSV
	 */
	exportCsv: async (formId: string) => {
		// This returns a file, so we need a different approach
		const token = localStorage.getItem('token');
		const response = await fetch(`/api/v1/forms/${formId}/export/csv`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {}
		});
		return response;
	}
};
