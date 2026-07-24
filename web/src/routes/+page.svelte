<script lang="ts">
	import { isAuthenticated } from '$lib/stores/auth';
	import { _ } from '$lib/i18n';
</script>

<svelte:head>
	<title>{$_('landing.metaTitle')}</title>
	<meta name="description" content={$_('landing.metaDescription')} />
</svelte:head>

<div class="min-h-screen bg-surface">
	<!-- Navbar -->
	<header class="fixed top-0 left-0 right-0 z-50 bg-surface/80 backdrop-blur-lg border-b border-neutral-100">
		<nav class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
			<a href="/" class="flex items-center gap-2">
				<div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
					<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 19v-8.93a2 2 0 01.89-1.664l7-4.666a2 2 0 012.22 0l7 4.666A2 2 0 0121 10.07V19M3 19a2 2 0 002 2h14a2 2 0 002-2M3 19l6.75-4.5M21 19l-6.75-4.5M3 10l6.75 4.5M21 10l-6.75 4.5m0 0l-1.14.76a2 2 0 01-2.22 0l-1.14-.76" />
					</svg>
				</div>
				<span class="text-xl font-bold text-neutral-900">OpenRSVP</span>
			</a>
			<div class="flex items-center gap-3">
				{#if $isAuthenticated}
					<a
						href="/events"
						class="rounded-md bg-primary px-5 py-2 text-sm font-semibold text-white hover:bg-primary-hover transition-colors shadow-sm"
					>
						{$_('nav.dashboard')}
					</a>
				{:else}
					<a
						href="/auth/login"
						class="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors px-3 py-2"
					>
						{$_('common.signIn')}
					</a>
					<a
						href="/auth/login"
						class="rounded-md bg-primary px-5 py-2 text-sm font-semibold text-white hover:bg-primary-hover transition-colors shadow-sm"
					>
						{$_('common.getStarted')}
					</a>
				{/if}
			</div>
		</nav>
	</header>

	<main>
		<!-- Hero Section -->
		<section class="relative overflow-hidden pt-16">
			<!-- Decorative background -->
			<div class="absolute inset-0 overflow-hidden" aria-hidden="true">
				<div class="absolute -top-40 -right-40 w-[600px] h-[600px] rounded-full bg-gradient-to-br from-primary-light via-pink-50 to-amber-50 opacity-60 blur-3xl"></div>
				<div class="absolute -bottom-20 -left-40 w-[500px] h-[500px] rounded-full bg-gradient-to-tr from-rose-100 via-primary-lighter to-amber-50 opacity-50 blur-3xl"></div>
				<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[400px] bg-gradient-to-r from-primary-lighter/0 via-primary-lighter/40 to-primary-lighter/0 rotate-12"></div>
			</div>

			<div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32 lg:py-40 text-center">
				<div class="inline-flex items-center gap-2 rounded-full bg-primary-light border border-primary-light px-4 py-1.5 text-sm font-medium text-primary mb-8">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					{$_('landing.badge')}
				</div>

				<h1 class="font-display text-5xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight text-neutral-900 mb-6 leading-[1.1]">
					{$_('landing.heroTitleLine1')}<br />
					<span class="bg-gradient-to-r from-primary via-pink-500 to-amber-500 bg-clip-text text-transparent">{$_('landing.heroTitleHighlight')}</span>
				</h1>

				<p class="text-lg sm:text-xl text-neutral-600 max-w-2xl mx-auto mb-10 leading-relaxed">
					{$_('landing.heroSubtitle')}
				</p>

				<div class="flex flex-col sm:flex-row items-center justify-center gap-4">
					<a
						href="/auth/login"
						class="group inline-flex items-center gap-2 rounded-lg bg-primary px-8 py-3.5 text-lg font-semibold text-white hover:bg-primary-hover transition-all shadow-lg shadow-primary/25 hover:shadow-primary/40"
					>
						{$_('landing.createFirstEvent')}
						<svg class="w-5 h-5 transition-transform group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
						</svg>
					</a>
					<a
						href="https://github.com/yannkr/openrsvp"
						target="_blank"
						rel="noopener noreferrer"
						class="inline-flex items-center gap-2 rounded-lg bg-surface px-8 py-3.5 text-lg font-semibold text-neutral-700 border border-neutral-200 hover:border-neutral-300 hover:bg-neutral-50 transition-all"
					>
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
						</svg>
						{$_('landing.viewOnGithub')}
					</a>
				</div>
			</div>
		</section>

		<!-- Features Grid -->
		<section class="relative py-24 sm:py-32">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="text-center mb-16">
					<h2 class="font-display text-3xl sm:text-4xl font-bold text-neutral-900 mb-4">
						{$_('landing.featuresTitle')}
					</h2>
					<p class="text-lg text-neutral-600 max-w-2xl mx-auto">
						{$_('landing.featuresSubtitle')}
					</p>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
					<!-- Beautiful Templates -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-primary to-rose-500 flex items-center justify-center mb-5 shadow-lg shadow-primary/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.templatesTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.templatesBody')}
						</p>
					</div>

					<!-- Privacy First -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-500 flex items-center justify-center mb-5 shadow-lg shadow-emerald-500/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.privacyTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.privacyBody')}
						</p>
					</div>

					<!-- Easy RSVPs -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-pink-500 to-rose-500 flex items-center justify-center mb-5 shadow-lg shadow-pink-500/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.rsvpsTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.rsvpsBody')}
						</p>
					</div>

					<!-- Smart Notifications -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center mb-5 shadow-lg shadow-amber-500/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.notificationsTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.notificationsBody')}
						</p>
					</div>

					<!-- Real-time Stats -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-sky-500 to-blue-500 flex items-center justify-center mb-5 shadow-lg shadow-sky-500/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.statsTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.statsBody')}
						</p>
					</div>

					<!-- Open Source -->
					<div class="group relative bg-surface rounded-xl p-8 shadow-sm border border-neutral-100 hover:shadow-md hover:border-primary-light transition-all">
						<div class="w-12 h-12 rounded-lg bg-gradient-to-br from-violet-500 to-purple-500 flex items-center justify-center mb-5 shadow-lg shadow-violet-500/25">
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-neutral-900 mb-2">{$_('landing.features.openSourceTitle')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.features.openSourceBody')}
						</p>
					</div>
				</div>
			</div>
		</section>

		<!-- How It Works -->
		<section class="relative py-24 sm:py-32 bg-neutral-50">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
				<div class="text-center mb-16">
					<h2 class="font-display text-3xl sm:text-4xl font-bold text-neutral-900 mb-4">
						{$_('landing.howTitle')}
					</h2>
					<p class="text-lg text-neutral-600 max-w-2xl mx-auto">
						{$_('landing.howSubtitle')}
					</p>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-8 lg:gap-12">
					<!-- Step 1 -->
					<div class="relative text-center">
						<div class="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-primary text-white text-2xl font-bold mb-6 shadow-lg shadow-primary/30">
							1
						</div>
						<h3 class="text-xl font-semibold text-neutral-900 mb-3">{$_('landing.step1Title')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.step1Body')}
						</p>
						<!-- Connector arrow (hidden on mobile) -->
						<div class="hidden md:block absolute top-8 left-[calc(100%-1rem)] w-8 text-neutral-300" aria-hidden="true">
							<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
							</svg>
						</div>
					</div>

					<!-- Step 2 -->
					<div class="relative text-center">
						<div class="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-pink-500 text-white text-2xl font-bold mb-6 shadow-lg shadow-pink-500/30">
							2
						</div>
						<h3 class="text-xl font-semibold text-neutral-900 mb-3">{$_('landing.step2Title')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.step2Body')}
						</p>
						<!-- Connector arrow (hidden on mobile) -->
						<div class="hidden md:block absolute top-8 left-[calc(100%-1rem)] w-8 text-neutral-300" aria-hidden="true">
							<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
								<path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
							</svg>
						</div>
					</div>

					<!-- Step 3 -->
					<div class="text-center">
						<div class="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-amber-500 text-white text-2xl font-bold mb-6 shadow-lg shadow-amber-500/30">
							3
						</div>
						<h3 class="text-xl font-semibold text-neutral-900 mb-3">{$_('landing.step3Title')}</h3>
						<p class="text-neutral-600 leading-relaxed">
							{$_('landing.step3Body')}
						</p>
					</div>
				</div>
			</div>
		</section>

		<!-- CTA Section -->
		<section class="relative py-24 sm:py-32 overflow-hidden">
			<div class="absolute inset-0" aria-hidden="true">
				<div class="absolute inset-0 bg-gradient-to-br from-primary via-pink-500 to-amber-500"></div>
				<div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg%20width%3D%2260%22%20height%3D%2260%22%20viewBox%3D%220%200%2060%2060%22%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%3E%3Cg%20fill%3D%22none%22%20fill-rule%3D%22evenodd%22%3E%3Cg%20fill%3D%22%23ffffff%22%20fill-opacity%3D%220.05%22%3E%3Cpath%20d%3D%22M36%2034v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6%2034v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6%204V0H4v4H0v2h4v4h2V6h4V4H6z%22%2F%3E%3C%2Fg%3E%3C%2Fg%3E%3C%2Fsvg%3E')] opacity-30"></div>
			</div>
			<div class="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
				<h2 class="font-display text-3xl sm:text-4xl font-bold text-white mb-6">
					{$_('landing.ctaTitle')}
				</h2>
				<p class="text-lg text-rose-100 mb-10 max-w-2xl mx-auto">
					{$_('landing.ctaSubtitle')}
				</p>
				<a
					href="/auth/login"
					class="inline-flex items-center gap-2 rounded-lg bg-surface px-8 py-3.5 text-lg font-semibold text-primary hover:bg-primary-lighter transition-colors shadow-lg"
				>
					{$_('landing.getStartedFree')}
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
					</svg>
				</a>
			</div>
		</section>
	</main>

	<!-- Footer -->
	<footer class="bg-neutral-900 text-neutral-400 py-12">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex flex-col sm:flex-row items-center justify-between gap-6">
				<div class="flex items-center gap-3">
					<div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
						<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 19v-8.93a2 2 0 01.89-1.664l7-4.666a2 2 0 012.22 0l7 4.666A2 2 0 0121 10.07V19M3 19a2 2 0 002 2h14a2 2 0 002-2M3 19l6.75-4.5M21 19l-6.75-4.5M3 10l6.75 4.5M21 10l-6.75 4.5m0 0l-1.14.76a2 2 0 01-2.22 0l-1.14-.76" />
						</svg>
					</div>
					<div>
						<div class="text-white font-semibold">OpenRSVP</div>
						<div class="text-sm">{$_('landing.footerTagline')}</div>
					</div>
				</div>
				<div class="flex items-center gap-6">
					<a
						href="https://github.com/yannkr/openrsvp"
						target="_blank"
						rel="noopener noreferrer"
						class="hover:text-white transition-colors inline-flex items-center gap-2 text-sm"
					>
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
						</svg>
						GitHub
					</a>
					<span class="text-sm">&copy; {new Date().getFullYear()} OpenRSVP</span>
				</div>
			</div>
		</div>
	</footer>
</div>
