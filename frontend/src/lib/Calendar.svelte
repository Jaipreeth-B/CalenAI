<script>
	import { tasks, selectedDate } from "./api.js";

	let now = new Date();
	let month = now.getMonth();
	let year = now.getFullYear();

	const daysOfWeek = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

	$: daysInMonth = new Date(year, month + 1, 0).getDate();
	$: firstDayOfMonth = new Date(year, month, 1).getDay();

	$: calendarDays = Array.from(
		{ length: firstDayOfMonth },
		() => null,
	).concat(Array.from({ length: daysInMonth }, (_, i) => i + 1));

	function hasTask(day) {
		if (!day) return false;
		return $tasks.some((t) => {
			if (!t.start_date_time) return false;
			const d = new Date(t.start_date_time);
			return (
				d.getDate() === day &&
				d.getMonth() === month &&
				d.getFullYear() === year
			);
		});
	}

	function changeMonth(delta) {
		month += delta;
		if (month < 0) {
			month = 11;
			year--;
		} else if (month > 11) {
			month = 0;
			year++;
		}
	}

	function selectDate(day) {
		if (day) {
			selectedDate.set(new Date(year, month, day));
		}
	}

	function goToToday() {
		const today = new Date();
		month = today.getMonth();
		year = today.getFullYear();
		selectedDate.set(new Date(year, month, today.getDate()));
	}

	function isSelected(day) {
		if (!day) return false;
		const d = new Date(year, month, day);
		return (
			d.getFullYear() === $selectedDate.getFullYear() &&
			d.getMonth() === $selectedDate.getMonth() &&
			d.getDate() === $selectedDate.getDate()
		);
	}

	function isToday(day) {
		if (!day) return false;
		const d = new Date(year, month, day);
		const t = new Date();
		return (
			d.getFullYear() === t.getFullYear() &&
			d.getMonth() === t.getMonth() &&
			d.getDate() === t.getDate()
		);
	}
</script>

<div class="flex flex-col h-full bg-white">
	<div
		class="px-8 py-5 flex justify-between items-center border-b border-gray-200"
	>
		<div class="flex flex-col">
			<h2
				class="text-[14px] font-bold text-gray-900 tracking-tight"
			>
				{new Date(year, month).toLocaleString(
					"default",
					{
						month: "long",
						year: "numeric",
					},
				)}
			</h2>
			<button
				on:click={goToToday}
				class="text-[10px] font-bold text-gray-400 hover:text-gray-900 transition-all text-left uppercase tracking-widest mt-0.5 cursor-pointer"
			>
				Today
			</button>
		</div>
		<div class="flex gap-2">
			<button
				on:click={() => changeMonth(-1)}
				class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-gray-900 transition-all cursor-pointer"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					><path d="m15 18-6-6 6-6" /></svg
				>
			</button>
			<button
				on:click={() => changeMonth(1)}
				class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-gray-900 transition-all cursor-pointer"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					><path d="m9 18 6-6-6-6" /></svg
				>
			</button>
		</div>
	</div>

	<div class="flex-1 overflow-y-auto px-6 py-6">
		<div class="grid grid-cols-7 gap-y-4">
			{#each daysOfWeek as day}
				<div
					class="text-center text-[11px] font-medium text-gray-500 mb-2"
				>
					{day[0]}
				</div>
			{/each}

			{#each calendarDays as day}
				<!-- svelte-ignore a11y-click-events-have-key-events -->
				<div
					class="aspect-square relative flex items-center justify-center {day
						? 'cursor-pointer group'
						: ''}"
					on:click={() => selectDate(day)}
				>
					{#if day}
						<div
							class="flex flex-col items-center gap-1.5"
						>
							<span
								class="text-[14px] font-medium w-8 h-8 rounded-full flex items-center justify-center transition-colors {isSelected(
									day,
								)
									? 'bg-gray-900 text-white shadow-sm'
									: isToday(
												day,
										  )
										? 'text-gray-900 font-bold border border-gray-300'
										: 'text-gray-700 group-hover:bg-gray-100'}"
							>
								{day}
							</span>
							{#if hasTask(day)}
								<div
									class="w-1.5 h-1.5 bg-gray-400 rounded-full {isSelected(
										day,
									)
										? 'bg-gray-900'
										: ''}"
								></div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.overflow-y-auto::-webkit-scrollbar {
		width: 0px;
	}
</style>
