package fontconfig

import (
	"strings"
	"unicode"
)

func FcStrCmpIgnoreCase(s1, s2 string) int {
	return strings.Compare(strings.ToLower(s1), strings.ToLower(s2))
}

func FcStrCmpIgnoreBlanksAndCase(s1, s2 string) int {
	return strings.Compare(ignoreBlanksAndCase(s1), ignoreBlanksAndCase(s2))
}

// Returns the location of `substr` in  `s`, ignoring case.
// Returns -1 if `substr` is not present in `s`.
func FcStrStrIgnoreCase(s, substr string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substr))
}

// The bulk of the time in FcFontMatch and FcFontSort goes to
// walking long lists of family names. We speed this up with a
// hash table.
type FamilyEntry struct {
	strong_value float64
	weak_value   float64
}

// map with strings key, ignoring blank and case
type FcHashTable map[string]*FamilyEntry

func ignoreBlanksAndCase(s string) string {
	s = strings.ToLower(s)
	return strings.TrimFunc(s, unicode.IsSpace)
}

func (h FcHashTable) lookup(s string) (*FamilyEntry, bool) {
	s = ignoreBlanksAndCase(s)
	e, ok := h[s]
	return e, ok
}

func (h FcHashTable) add(s string, v *FamilyEntry) {
	s = ignoreBlanksAndCase(s)
	h[s] = v
}

// IgnoreBlanksAndCase
type familyBlankHash map[string]int

func (h familyBlankHash) lookup(s string) (int, bool) {
	s = ignoreBlanksAndCase(s)
	e, ok := h[s]
	return e, ok
}

func (h familyBlankHash) add(s string, v int) {
	s = ignoreBlanksAndCase(s)
	h[s] = v
}

func (h familyBlankHash) del(s string) {
	s = ignoreBlanksAndCase(s)
	delete(h, s)
}

// IgnoreCase
type familyHash map[string]int

func (h familyHash) lookup(s string) (int, bool) {
	s = strings.ToLower(s)
	e, ok := h[s]
	return e, ok
}

func (h familyHash) add(s string, v int) {
	s = strings.ToLower(s)
	h[s] = v
}

func (h familyHash) del(s string) {
	s = strings.ToLower(s)
	delete(h, s)
}

// const FC_HASH_SIZE = 227

//  typedef struct _FcHashBucket {
// 	 struct _FcHashBucket  *next;
// 	 void                  *key;
// 	 void                  *value;
//  } FcHashBucket;

//  struct _FcHashTable {
// 	 FcHashBucket  *buckets[FC_HASH_SIZE];
// 	 FcHashFunc     hash_func;
// 	 FcCompareFunc  compare_func;
// 	 FcCopyFunc     key_copy_func;
// 	 FcCopyFunc     value_copy_func;
// 	 FcDestroyFunc  key_destroy_func;
// 	 FcDestroyFunc  value_destroy_func;
//  };

//  FcBool
//  FcHashStrCopy (const void  *src,
// 			void       **dest)
//  {
// 	 *dest = FcStrdup (src);

// 	 return *dest != NULL;
//  }

//  FcHashTable *
//  FcHashTableCreate (FcHashFunc    hash_func,
// 			FcCompareFunc compare_func,
// 			FcCopyFunc    key_copy_func,
// 			FcCopyFunc    value_copy_func,
// 			FcDestroyFunc key_destroy_func,
// 			FcDestroyFunc value_destroy_func)
//  {
// 	 FcHashTable *ret = malloc (sizeof (FcHashTable));

// 	 if (ret)
// 	 {
// 	 memset (ret->buckets, 0, sizeof (FcHashBucket *) * FC_HASH_SIZE);
// 	 ret->hash_func = hash_func;
// 	 ret->compare_func = compare_func;
// 	 ret->key_copy_func = key_copy_func;
// 	 ret->value_copy_func = value_copy_func;
// 	 ret->key_destroy_func = key_destroy_func;
// 	 ret->value_destroy_func = value_destroy_func;
// 	 }
// 	 return ret;
//  }

//  void
//  FcHashTableDestroy (FcHashTable *table)
//  {
// 	 int i;

// 	 for (i = 0; i < FC_HASH_SIZE; i++)
// 	 {
// 	 FcHashBucket *bucket = table->buckets[i], *prev;

// 	 while (bucket)
// 	 {
// 		 if (table->key_destroy_func)
// 		 table->key_destroy_func (bucket->key);
// 		 if (table->value_destroy_func)
// 		 table->value_destroy_func (bucket->value);
// 		 prev = bucket;
// 		 bucket = bucket->next;
// 		 free (prev);
// 	 }
// 	 table->buckets[i] = NULL;
// 	 }
// 	 free (table);
//  }

//  FcBool
//  FcHashTableFind (FcHashTable  *table,
// 		  const void   *key,
// 		  void        **value)
//  {
// 	 FcHashBucket *bucket;
// 	 FcChar32 hash = table->hash_func (key);

// 	 for (bucket = table->buckets[hash % FC_HASH_SIZE]; bucket; bucket = bucket->next)
// 	 {
// 	 if (!table->compare_func(bucket->key, key))
// 	 {
// 		 if (table->value_copy_func)
// 		 {
// 		 if (!table->value_copy_func (bucket->value, value))
// 			 return FcFalse;
// 		 }
// 		 else
// 		 *value = bucket->value;
// 		 return FcTrue;
// 	 }
// 	 }
// 	 return FcFalse;
//  }

//  static FcBool
//  FcHashTableAddInternal (FcHashTable *table,
// 			 void        *key,
// 			 void        *value,
// 			 FcBool       replace)
//  {
// 	 FcHashBucket **prev, *bucket, *b;
// 	 FcChar32 hash = table->hash_func (key);
// 	 FcBool ret = FcFalse;

// 	 bucket = (FcHashBucket *) malloc (sizeof (FcHashBucket));
// 	 if (!bucket)
// 	 return FcFalse;
// 	 memset (bucket, 0, sizeof (FcHashBucket));
// 	 if (table->key_copy_func)
// 	 ret |= !table->key_copy_func (key, &bucket->key);
// 	 else
// 	 bucket->key = key;
// 	 if (table->value_copy_func)
// 	 ret |= !table->value_copy_func (value, &bucket->value);
// 	 else
// 	 bucket->value = value;
// 	 if (ret)
// 	 {
// 	 destroy:
// 	 if (bucket->key && table->key_destroy_func)
// 		 table->key_destroy_func (bucket->key);
// 	 if (bucket->value && table->value_destroy_func)
// 		 table->value_destroy_func (bucket->value);
// 	 free (bucket);

// 	 return !ret;
// 	 }
//    retry:
// 	 for (prev = &table->buckets[hash % FC_HASH_SIZE];
// 	  (b = fc_atomic_ptr_get (prev)); prev = &(b->next))
// 	 {
// 	 if (!table->compare_func (b->key, key))
// 	 {
// 		 if (replace)
// 		 {
// 		 bucket->next = b->next;
// 		 if (!fc_atomic_ptr_cmpexch (prev, b, bucket))
// 			 goto retry;
// 		 bucket = b;
// 		 }
// 		 else
// 		 ret = FcTrue;
// 		 goto destroy;
// 	 }
// 	 }
// 	 bucket->next = NULL;
// 	 if (!fc_atomic_ptr_cmpexch (prev, b, bucket))
// 	 goto retry;

// 	 return FcTrue;
//  }

//  FcBool
//  FcHashTableAdd (FcHashTable *table,
// 		 void        *key,
// 		 void        *value)
//  {
// 	 return FcHashTableAddInternal (table, key, value, FcFalse);
//  }

//  FcBool
//  FcHashTableReplace (FcHashTable *table,
// 			 void        *key,
// 			 void        *value)
//  {
// 	 return FcHashTableAddInternal (table, key, value, FcTrue);
//  }

//  FcBool
//  FcHashTableRemove (FcHashTable *table,
// 			void        *key)
//  {
// 	 FcHashBucket **prev, *bucket;
// 	 FcChar32 hash = table->hash_func (key);
// 	 FcBool ret = FcFalse;

//  retry:
// 	 for (prev = &table->buckets[hash % FC_HASH_SIZE];
// 	  (bucket = fc_atomic_ptr_get (prev)); prev = &(bucket->next))
// 	 {
// 	 if (!table->compare_func (bucket->key, key))
// 	 {
// 		 if (!fc_atomic_ptr_cmpexch (prev, bucket, bucket->next))
// 		 goto retry;
// 		 if (table->key_destroy_func)
// 		 table->key_destroy_func (bucket->key);
// 		 if (table->value_destroy_func)
// 		 table->value_destroy_func (bucket->value);
// 		 free (bucket);
// 		 ret = FcTrue;
// 		 break;
// 	 }
// 	 }

// 	 return ret;
//  }
