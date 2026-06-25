#![cfg_attr(not(feature = "std"), no_std)]

/// Pallet Affinity: The Protein-Interaction Graph and SRRP Routing substrate.
/// This pallet tracks the structural affinity between agents, which governs team formation
/// and routing topologies without explicit hop counts.

pub use pallet::*;

#[frame_support::pallet]
pub mod pallet {
	use frame_support::pallet_prelude::*;
	use frame_system::pallet_prelude::*;

	#[pallet::pallet]
	pub struct Pallet<T>(_);

	#[pallet::config]
	pub trait Config: frame_system::Config {
		type RuntimeEvent: From<Event<Self>> + IsType<<Self as frame_system::Config>::RuntimeEvent>;
	}

	/// Maps an interaction edge between two agents to an AffinityScore.
	/// Ordered tuples are used to ensure bidirectional interaction is scored uniformly.
	#[pallet::storage]
	#[pallet::getter(fn interaction_graph)]
	pub(super) type InteractionGraph<T: Config> = StorageDoubleMap<
		_,
		Blake2_128Concat, T::AccountId,
		Blake2_128Concat, T::AccountId,
		u32, // Affinity Score
		OptionQuery
	>;

	#[pallet::event]
	#[pallet::generate_deposit(pub(super) fn deposit_event)]
	pub enum Event<T: Config> {
		/// A new affinity bond was formed or strengthened between two agents.
		AffinityBondUpdated { agent_a: T::AccountId, agent_b: T::AccountId, new_score: u32 },
		/// A protein complex team was crystallized (formed).
		ProteinComplexFormed { team_id: [u8; 16], members: Vec<T::AccountId> },
	}

	#[pallet::error]
	pub enum Error<T> {
		CannotBondWithSelf,
		MaxAffinityReached,
	}

	#[pallet::call]
	impl<T: Config> Pallet<T> {
		/// Records an interaction between two agents, increasing their affinity score.
		/// Over time, agents with high affinity crystallize into functional "Protein Complex" teams.
		#[pallet::call_index(0)]
		#[pallet::weight(10_000 + T::DbWeight::get().reads_writes(1, 1).ref_time())]
		pub fn record_interaction(
			origin: OriginFor<T>,
			partner: T::AccountId,
			interaction_strength: u32
		) -> DispatchResult {
			let who = ensure_signed(origin)?;

			ensure!(who != partner, Error::<T>::CannotBondWithSelf);

			// Deterministic ordering to ensure A <-> B is the same edge as B <-> A
			let (agent_a, agent_b) = if who < partner { (who.clone(), partner.clone()) } else { (partner.clone(), who.clone()) };

			let current_score = InteractionGraph::<T>::get(&agent_a, &agent_b).unwrap_or(0);
			let new_score = current_score.saturating_add(interaction_strength);

			ensure!(new_score < 100_000, Error::<T>::MaxAffinityReached);

			InteractionGraph::<T>::insert(&agent_a, &agent_b, new_score);

			Self::deposit_event(Event::AffinityBondUpdated { agent_a, agent_b, new_score });

			Ok(())
		}
	}
}
