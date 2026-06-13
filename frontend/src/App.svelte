<script>
	import { onMount } from "svelte";
	import Sidebar from "./lib/Sidebar.svelte";
	import Chat from "./lib/Chat.svelte";
	import Calendar from "./lib/Calendar.svelte";
	import TaskList from "./lib/TaskList.svelte";
	import { fetchTasks, fetchSessions } from "./lib/api.js";
	import { fade } from "svelte/transition";

	onMount(() => {
		fetchTasks();
		fetchSessions();
	});
</script>

<main
	class="flex h-screen w-full bg-white overflow-hidden text-gray-900 selection:bg-gray-900 selection:text-white"
	in:fade={{ duration: 600 }}
>
	<!-- Left Sidebar: Chat Sessions -->
	<aside
		class="w-[260px] h-full flex-shrink-0 bg-gray-50 border-r border-gray-200"
	>
		<Sidebar />
	</aside>

	<!-- Center: Main Chat Area -->
	<section
		class="w-[550px] h-full flex-shrink-0 border-r border-gray-200 bg-white"
	>
		<Chat />
	</section>

	<!-- Right: Calendar & Tasks -->
	<section class="flex-1 flex flex-col h-full overflow-hidden bg-white">
		<!-- Calendar View -->
		<div class="h-[45%] overflow-hidden border-b border-gray-200">
			<Calendar />
		</div>

		<!-- Task List View -->
		<div class="flex-1 overflow-y-auto">
			<TaskList />
		</div>
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		background-color: white;
	}

	:global(.overflow-y-auto::-webkit-scrollbar) {
		width: 0px;
	}
</style>
