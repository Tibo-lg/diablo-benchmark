module counter::counter {
    use sui::transfer;
    use sui::object::{Self, UID};
    use sui::tx_context::{Self, TxContext};

    struct Counter has key {
        id: UID,
        i: u64,
    }

    fun init(ctx: &mut TxContext) {
        let newCounter = Counter {
            id: object::new(ctx),
            i: 0,
        };

        transfer::transfer(newCounter, tx_context::sender(ctx))
    }

    entry public fun increment(counter: &mut Counter) {
        counter.i = counter.i + 1
    }
}
