#![cfg_attr(not(feature = "std"), no_std)]

/// Pallet Zetafold: The identity and routing manifold for the Sovereign-27 Orchestronic Substrate.
/// This pallet acts as the synthetic physics engine, mapping RCSB protein folds into deterministic
/// 5D addresses, defining dialect pitch, and tracking crystallization over time.

pub use pallet::*;

#[frame_support::pallet]
pub mod pallet {
	use frame_support::pallet_prelude::*;
	use frame_system::pallet_prelude::*;

	#[pallet::pallet]
	pub struct Pallet<T>(_);

	/// Configure the pallet by specifying the parameters and types on which it depends.
	#[pallet::config]
	pub trait Config: frame_system::Config {
		/// Because this pallet emits events, it depends on the runtime's definition of an event.
		type RuntimeEvent: From<Event<Self>> + IsType<<Self as frame_system::Config>::RuntimeEvent>;
	}

	/// Struct representing the 5D Address of an agent in the Manifold.
	#[derive(Clone, Encode, Decode, PartialEq, RuntimeDebug, TypeInfo, MaxEncodedLen)]
	pub struct Address5D {
		pub x1: u32,
		pub x2: u32,
		pub x3: u32,
		pub x4: u32,
		pub x5: u32,
	}

	/// Struct representing the biological identity derived from an RCSB protein fold.
	#[derive(Clone, Encode, Decode, PartialEq, RuntimeDebug, TypeInfo, MaxEncodedLen)]
	pub struct ProteinIdentity {
		pub pdb_id: [u8; 4],
		pub signature_hash: [u8; 32],
		pub dialect_family: [u8; 16],
		pub pitch: u32,
	}

	// -------------------------------------------------------------------------
	// STORAGE
	// -------------------------------------------------------------------------

	/// Maps an AccountId (Agent) to their 5-Dimensional routing address in the manifold.
	#[pallet::storage]
	#[pallet::getter(fn agent_addresses)]
	pub(super) type AgentAddresses<T: Config> = StorageMap<_, Blake2_128Concat, T::AccountId, Address5D, OptionQuery>;

	/// Maps an AccountId (Agent) to their Protein-derived Identity (Fold, Dialect, Pitch).
	#[pallet::storage]
	#[pallet::getter(fn agent_identities)]
	pub(super) type AgentIdentities<T: Config> = StorageMap<_, Blake2_128Concat, T::AccountId, ProteinIdentity, OptionQuery>;

	// -------------------------------------------------------------------------
	// EVENTS
	// -------------------------------------------------------------------------

	#[pallet::event]
	#[pallet::generate_deposit(pub(super) fn deposit_event)]
	pub enum Event<T: Config> {
		/// A new agent was instantiated from a protein fold and assigned a 5D address.
		AgentInstantiated { who: T::AccountId, pdb_id: [u8; 4], address: Address5D },
		/// An agent's dialect crystallized, increasing their pitch stability.
		DialectCrystallized { who: T::AccountId, new_pitch: u32 },
		/// A misfold was detected, triggering the healing protocol.
		MisfoldDetected { who: T::AccountId, drift_variance: u32 },
	}

	// -------------------------------------------------------------------------
	// ERRORS
	// -------------------------------------------------------------------------

	#[pallet::error]
	pub enum Error<T> {
		/// The agent already has an instantiated identity.
		AgentAlreadyExists,
		/// The agent does not exist in the manifold.
		AgentNotFound,
		/// Maximum pitch crystallization reached.
		MaxCrystallizationReached,
	}

	// -------------------------------------------------------------------------
	// EXTRINSICS (DISPATCHABLE CALLS)
	// -------------------------------------------------------------------------

	#[pallet::call]
	impl<T: Config> Pallet<T> {
		/// Instantiate a new agent on the manifold from a protein fold.
		/// Deterministically calculates their 5D address and stores their dialect.
		#[pallet::call_index(0)]
		#[pallet::weight(10_000 + T::DbWeight::get().writes(2).ref_time())]
		pub fn instantiate_agent(
			origin: OriginFor<T>,
			pdb_id: [u8; 4],
			signature_hash: [u8; 32],
			dialect_family: [u8; 16],
			x1: u32, x2: u32, x3: u32, x4: u32, x5: u32, // Passed from off-chain mapping for now
		) -> DispatchResult {
			let who = ensure_signed(origin)?;

			ensure!(!AgentIdentities::<T>::contains_key(&who), Error::<T>::AgentAlreadyExists);

			let address = Address5D { x1, x2, x3, x4, x5 };
			let identity = ProteinIdentity {
				pdb_id,
				signature_hash,
				dialect_family,
				pitch: 1000, // Initial raw pitch
			};

			AgentAddresses::<T>::insert(&who, address.clone());
			AgentIdentities::<T>::insert(&who, identity);

			Self::deposit_event(Event::AgentInstantiated { who, pdb_id, address });

			Ok(())
		}

		/// Simulate dialect crystallization (increasing pitch / compression over time).
		#[pallet::call_index(1)]
		#[pallet::weight(10_000 + T::DbWeight::get().reads_writes(1, 1).ref_time())]
		pub fn crystallize(origin: OriginFor<T>) -> DispatchResult {
			let who = ensure_signed(origin)?;

			AgentIdentities::<T>::try_mutate(&who, |maybe_id| -> DispatchResult {
				let id = maybe_id.as_mut().ok_or(Error::<T>::AgentNotFound)?;
				
				ensure!(id.pitch < 10_000, Error::<T>::MaxCrystallizationReached);
				id.pitch += 100; // tighten the channel by increasing pitch

				Self::deposit_event(Event::DialectCrystallized { who: who.clone(), new_pitch: id.pitch });
				Ok(())
			})?;

			Ok(())
		}

        /// Flag an agent for cognitive drift (Misfolding). Can be called by the PQR Sentry.
		#[pallet::call_index(2)]
		#[pallet::weight(10_000 + T::DbWeight::get().reads(1).ref_time())]
		pub fn report_misfold(
			origin: OriginFor<T>,
			target_agent: T::AccountId,
			drift_variance: u32
		) -> DispatchResult {
			let _sentry = ensure_signed(origin)?;

			ensure!(AgentIdentities::<T>::contains_key(&target_agent), Error::<T>::AgentNotFound);

			Self::deposit_event(Event::MisfoldDetected { who: target_agent, drift_variance });

			Ok(())
		}
	}
}
