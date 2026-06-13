<script>
	import { tasks, toggleTask, deleteTask, selectedDate } from "./api.js";
	import { flip } from "svelte/animate";
	import { fade } from "svelte/transition";

	function formatDate(dateStr) {
		if (!dateStr) return "";
		const date = new Date(dateStr);
		return date.toLocaleTimeString([], {
			hour: "2-digit",
			minute: "2-digit",
		});
	}

	// Helper to compare dates without time/timezone shifts
	function isSameDay(date1, date2) {
		if (!date1 || !date2) return false;
		const d1 = new Date(date1);
		const d2 = new Date(date2);
		return (
			d1.getFullYear() === d2.getFullYear() &&
			d1.getMonth() === d2.getMonth() &&
			d1.getDate() === d2.getDate()
		);
	}

	// Filter tasks: scheduled for today OR unscheduled (if today is today)
	$: isToday = isSameDay(new Date(), $selectedDate);

	$: scheduledTasks = $tasks.filter((task) =>
		isSameDay(task.start_date_time, $selectedDate),
	);

	$: unscheduledTasks = $tasks.filter((task) => !task.start_date_time);

	$: displayTasks = isToday
		? [...scheduledTasks, ...unscheduledTasks]
		: scheduledTasks;

	$: sortedTasks = [...displayTasks].sort((a, b) => {
		if (a.is_completed === b.is_completed) {
			if (!a.start_date_time) return 1;
			if (!b.start_date_time) return -1;
			return (
				new Date(a.start_date_time) -
				new Date(b.start_date_time)
			);
		}
		return a.is_completed ? 1 : -1;
	});

	$: formattedSelectedDate = $selectedDate.toLocaleDateString("default", {
		weekday: "long",
		month: "short",
		day: "numeric",
	});
</script>

<div class="flex flex-col h-full bg-white">
	<div
		class="px-8 py-5 border-b border-gray-200 flex justify-between items-center"
	>
		<div class="flex items-center gap-2">
			<h2 class="text-[13px] font-bold text-gray-800">
				Schedule
			</h2>
			{#if isToday}
				<span
					class="text-[9px] font-bold bg-gray-900 text-white px-1.5 py-0.5 rounded uppercase tracking-wider"
					>Today</span
				>
			{/if}
		</div>
		<span
			class="text-[11px] font-medium text-gray-500 bg-gray-100 px-2 py-1 rounded-md"
			>{formattedSelectedDate}</span
		>
	</div>

	<div class="flex-1 overflow-y-auto px-6 py-4 space-y-3">
		{#each sortedTasks as task (task.id)}
			<div
				animate:flip={{ duration: 400 }}
				in:fade={{ duration: 300 }}
				class="group flex items-center gap-4 p-4 rounded-xl border border-gray-200 hover:border-gray-300 transition-all {task.is_completed
					? 'opacity-50'
					: ''}"
			>
				<button
					on:click={() => toggleTask(task)}
					class="w-5 h-5 rounded-full border flex items-center justify-center transition-all cursor-pointer {task.is_completed
						? 'bg-gray-900 border-gray-900 text-white'
						: 'bg-white border-gray-300 hover:border-gray-500'}"
				>
					{#if task.is_completed}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="12"
							height="12"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="3"
							stroke-linecap="round"
							stroke-linejoin="round"
							><path
								d="M20 6 9 17l-5-5"
							/></svg
						>
					{/if}
				</button>

				<div class="flex-1 min-w-0">
					<p
						class="text-[14px] font-medium text-gray-900 truncate {task.is_completed
							? 'line-through text-gray-500'
							: ''}"
					>
						{task.title}
					</p>
					{#if task.start_date_time}
						<p
							class="text-[11px] font-medium text-gray-500 mt-0.5"
						>
							{formatDate(task.start_date_time)}
							{#if task.end_date_time}
								- {formatDate(task.end_date_time)}
							{/if}
						</p>
					{/if}
				</div>

				<button
					on:click={() => deleteTask(task.id)}
					class="text-[12px] font-medium text-gray-400 hover:text-red-500 transition-all cursor-pointer opacity-0 group-hover:opacity-100 px-2"
				>
					Delete
				</button>
			</div>
		{:else}
			<div
				class="h-full flex flex-col items-center justify-center py-10 text-gray-400"
			>
				<p class="text-sm font-medium">
					No tasks scheduled.
				</p>
				<p
					class="text-[10px] uppercase tracking-widest mt-2 opacity-50"
				>
					Enjoy your day
				</p>
			</div>
		{/each}
	</div>
</div>

<style>
	.overflow-y-auto::-webkit-scrollbar {
		width: 0px;
	}
</style>